/**
 * Welcome to Cloudflare Workers!
 *
 * - Run `wrangler dev src/index.ts` in your terminal to start a development server
 * - Open a browser tab at http://localhost:8787/ to see your worker in action
 * - Run `wrangler publish src/index.ts --name my-worker` to publish your worker
 *
 * Learn more at https://developers.cloudflare.com/workers/
 */
import { parse } from 'cookie';
import { Toucan } from 'toucan-js';
import jwt, { JwtPayload } from '@tsndr/cloudflare-worker-jwt';
import { RequestInitCfProperties } from '@cloudflare/workers-types';

const WEBSOCKET_PASS_HEADERS = [
  'Accept-Encoding',
  'Accept-Language',
  'Connection',
  'Cookie',
  'Host',
  'Origin',
  'Pragma',
  'Sec-Websocket-Extensions',
  'Sec-Websocket-Protocol',
  'Sec-Websocket-Version',
  'Upgrade',
  'User-Agent',
].map((header) => header.toLowerCase());

const CACHE_FILE_EXTENSIONS: string[] = [
  'js',
  'css',
  'png',
  'jpg',
  'jpeg',
  'ico',
];

const ILO_SESSION_KEY_COOKIE: string = 'sessionKey';

export interface Env {
    JWT_SECRET: string;
    JWT_COOKIE: string;
    PROXY_PORT: number;
    SENTRY_DSN: string;
}

interface SessionData extends JwtPayload {
  ip: string;
  upstream: string;
}

// Returns the current unix timestamp
function unixTimestamp(date = Date.now()) {
  return Math.floor(date / 1000);
}

export default {
  async fetch(
    request: Request,
    env: Env,
    ctx: ExecutionContext,
  ): Promise<Response> {
    const sentry = new Toucan({
      dsn: env.SENTRY_DSN,
      release: '1.0.0',
      sampleRate: 1.0,
      attachStacktrace: true,
      autoSessionTracking: true,
      context: ctx,
      request,
    });

    const url = new URL(request.url);

    // move token to session cookie and redirect to /
    if (url.pathname === '/gpsession' && url.search != '') {
      const params = new URLSearchParams(url.search);
      if (params.get('token') != null) {
        url.pathname = '';
        url.search = '';
        return new Response(null, {
          status: 302,
          headers: {
            Location: url.toString(),
            'Set-Cookie': `${env.JWT_COOKIE} = ${params.get(
              'token',
            )}; Path=/; Secure; HttpOnly; SameSite=Lax`,
          },
        });
      }
    }

    sentry.setExtras({
      requestHeaders: request.headers,
    });

    const cookies = parse(request.headers.get('Cookie') || '');
    if (!cookies[env.JWT_COOKIE]) {
      return new Response(
        JSON.stringify({
          error: `no jwt cookie ${env.JWT_COOKIE} found`,
        }),
        {
          status: 403,
        },
      );
    }

    try {
      if (!(await jwt.verify(cookies[env.JWT_COOKIE], env.JWT_SECRET))) {
        return new Response(
          JSON.stringify({
            error: 'invalid jwt signature',
          }),
          {
            status: 403,
          },
        );
      }
    } catch (error) {
      return new Response(
        JSON.stringify({
          error: 'invalid jwt token',
        }),
        {
          status: 403,
        },
      );
    }

    const { payload } = jwt.decode(cookies[env.JWT_COOKIE]);
    const sessionData = payload as SessionData;
    const upstreamPort = url.port !== '' ? url.port : '443';

    const cf: RequestInitCfProperties = {
      cacheEverything: false,
      scrapeShield: false,
      polish: 'off',
      apps: false,
    };

    // websocket handling
    const upgradeHeader = request.headers.get('Upgrade');
    if (upgradeHeader && upgradeHeader.toLowerCase() === 'websocket') {
      const webSocketPair = new WebSocketPair();
      const client = webSocketPair[0],
        server = webSocketPair[1];

      const protocols: string[] = [];
      if (request.headers.get('Sec-Websocket-Protocol') != null) {
        const secWebsocketProtocol = request.headers.get(
          'Sec-Websocket-Protocol',
        ) as string;
        secWebsocketProtocol.split(',').forEach((protocol) => {
          protocols.push(protocol.trim());
        });
      }

      const passHeaders: Record<string, string> = {};
      request.headers.forEach((value, key) => {
        if (WEBSOCKET_PASS_HEADERS.includes(key.toLowerCase())) {
          console.log(`passing header ${key}: ${value}`);
          passHeaders[key] = value;
        }
      });

      let base64Headers = btoa(JSON.stringify(passHeaders));
      let websocketUrl = `wss://${sessionData.upstream}:${env.PROXY_PORT}/websocket/${sessionData.ip}/${upstreamPort}/${url.pathname}`;
      websocketUrl += `?headers=${base64Headers}`;
      const upstream = new WebSocket(
        websocketUrl,
        protocols.length > 0 ? protocols : undefined,
      );

      console.log({
        protocol: upstream.protocol,
        websocketUrl,
        protocols,
        base64Headers,
      });

      upstream.addEventListener('error', (event) => {
        sentry.captureMessage(event.message, 'debug', {
          data: event,
        });
        server.close(undefined, event.message);
        upstream.close(undefined, event.message);
      });

      server.addEventListener('error', (event) => {
        sentry.captureMessage(event.message, 'debug', {
          data: event,
        });
        server.close(undefined, event.message);
        upstream.close(undefined, event.message);
      });

      upstream.addEventListener('close', (event) => {
        sentry.addBreadcrumb({
          message: 'upstream websocket closed',
          data: {
            message: JSON.stringify(event),
          },
        });
        sentry.captureMessage('upstream websocket closed', 'debug');
        server.close(event.code, event.reason);
        client.close(event.code, event.reason);
      });

      server.addEventListener('close', (event) => {
        sentry.addBreadcrumb({
          message: 'server websocket closed',
          data: {
            message: JSON.stringify(event),
          },
        });
        sentry.captureMessage('server websocket closed', 'debug');
        upstream.close(event.code, event.reason);
        client.close(event.code, event.reason);
      });

      upstream.addEventListener('open', (event) => {
        sentry.addBreadcrumb({
          message: 'upstream websocket opened',
        });
        server.accept();
      });

      // when the upstream sends a message, forward it to the client
      upstream.addEventListener('message', (event) => {
        sentry.addBreadcrumb({
          message: 'message received from upstream',
          data: {
            event: JSON.stringify(event),
          },
        });
        server.send(event.data);
      });

      // when the client sends a message, forward it to the server, forward to upstream
      server.addEventListener('message', (event) => {
        sentry.addBreadcrumb({
          message: 'message received from client',
          data: {
            event: JSON.stringify(event),
          },
        });
        upstream.send(event.data);
      });

      const headers = new Headers();
      if (request.headers.has('Sec-WebSocket-Protocol')) {
        const secWebsocketProtocol = request.headers.get(
          'Sec-Websocket-Protocol',
        ) as string;
        if (secWebsocketProtocol.includes('binary')) {
          headers.set('Sec-WebSocket-Protocol', 'binary');
        } else {
          headers.set('Sec-WebSocket-Protocol', secWebsocketProtocol);
        }
      }
      headers.set('Connection', 'Upgrade');

      return new Response(null, {
        status: 101,
        webSocket: client,
        headers,
        cf,
      });
    }

    const proxyUrl = new URL(url);
    proxyUrl.host = sessionData.upstream;
    proxyUrl.port = env.PROXY_PORT.toString();

    console.log('upstream URL: ', proxyUrl.toString());

    const newRequest = new Request(proxyUrl, request);
    const requestHeaders: Map<string, string> = new Map(newRequest.headers);
    requestHeaders.set('X-Forwarded-Host', url.host);
    requestHeaders.set('X-Forwarded-Port', url.port);
    requestHeaders.set('X-Forwarded-Proto', url.protocol);
    requestHeaders.set('X-PM-Host', sessionData.ip);
    requestHeaders.set('X-PM-Port', upstreamPort);

    // Hacky workaround for the Redfish calls to the iLO
    if (cookies[ILO_SESSION_KEY_COOKIE]) {
      requestHeaders.set('X-PM-Token', cookies[ILO_SESSION_KEY_COOKIE].trim());
    }

    // cache files with CACHE_FILE_EXTENSIONS for one hour
    if (
      url.pathname.includes('.') &&
      CACHE_FILE_EXTENSIONS.includes(url.pathname.split('.').pop() as string)
    ) {
      // cacheTtl should be now - sessionData.exp
      cf.cacheTtl = sessionData.exp! - unixTimestamp();
      if (cf.cacheTtl > 3600) {
        cf.cacheTtl = 3600;
      }
      cf.polish = 'lossy';
      cf.minify = {
        javascript: true,
        css: true,
      };
      cf.cacheKey = `${url.protocol}//${sessionData.ip}${url.pathname}${url.search}`;
      console.log(`Cache key "${cf.cacheKey}" (ttl: ${cf.cacheTtl}).`);
    }

    console.log(JSON.stringify(requestHeaders.keys()));
    console.log({
      url,
      proxyUrl,
    });

    return fetch(newRequest, {
      headers: requestHeaders,
      cf,
    });
  },
};
