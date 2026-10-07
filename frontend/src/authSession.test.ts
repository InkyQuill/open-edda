import { beforeEach, describe, expect, it, vi } from 'vitest';
import { authFetch } from './api';
import { clearToken, getToken, subscribeSession, wasSessionExpired } from './authApi';

beforeEach(() => {
  const data = new Map<string,string>();
  vi.stubGlobal('localStorage', { getItem: (key:string) => data.get(key) ?? null, setItem: (key:string,value:string) => data.set(key,value), removeItem: (key:string) => data.delete(key) });
  clearToken();
});

describe('session rejection', () => {
  it('invalidates the rejected session and notifies the router', async () => {
    localStorage.setItem('open_edda_token','old');
    const listener = vi.fn(); const unsubscribe = subscribeSession(listener);
    vi.stubGlobal('fetch',vi.fn().mockResolvedValue(new Response('',{status:401})));
    await authFetch('/api/projects');
    expect(getToken()).toBeNull(); expect(wasSessionExpired()).toBe(true); expect(listener).toHaveBeenCalledOnce();
    unsubscribe();
  });
  it('does not expire a newer login when an older request returns late', async () => {
    localStorage.setItem('open_edda_token','old');
    vi.stubGlobal('fetch',vi.fn().mockImplementation(async () => { localStorage.setItem('open_edda_token','new'); return new Response('',{status:401}); }));
    await authFetch('/api/projects');
    expect(getToken()).toBe('new'); expect(wasSessionExpired()).toBe(false);
  });
  it('keeps the session for access denials and server failures', async () => {
    localStorage.setItem('open_edda_token','valid');
    for (const status of [403,500]) {
      vi.stubGlobal('fetch',vi.fn().mockResolvedValue(new Response('',{status})));
      await authFetch('/api/projects'); expect(getToken()).toBe('valid');
    }
  });
});

it('shares one refresh across concurrent requests and retries their bodies', async () => {
  localStorage.setItem('open_edda_token','old');
  localStorage.setItem('open_edda_session','session-one');
  let refreshes = 0;
  const calls: string[] = [];
  vi.stubGlobal('fetch',vi.fn().mockImplementation(async (path: string, init: RequestInit) => {
    if (path === '/api/auth/refresh') {
      refreshes++;
      await Promise.resolve();
      return Response.json({token:'renewed',refreshExpiresAt:Date.now()/1000+30*86400});
    }
    calls.push(String(init.body));
    return new Response('',{status:new Headers(init.headers).get('Authorization')==='Bearer renewed'?200:401});
  }));
  const results = await Promise.all([authFetch('/api/a',{method:'PUT',body:'draft'}),authFetch('/api/b',{method:'PUT',body:'draft'})]);
  expect(results.map(r=>r.status)).toEqual([200,200]);
  expect(refreshes).toBe(1); expect(calls).toEqual(['draft','draft','draft','draft']);
});

it('refreshes proactively in the last week and preserves login on a temporary refresh failure', async () => {
  localStorage.setItem('open_edda_token','valid');
  localStorage.setItem('open_edda_refresh_expires_at',String(Date.now()/1000+6*86400));
  const fetchMock=vi.fn().mockResolvedValue(new Response('',{status:503}));
  vi.stubGlobal('fetch',fetchMock);
  await expect(authFetch('/api/projects')).rejects.toThrow('Не удалось продлить');
  expect(getToken()).toBe('valid'); expect(fetchMock).toHaveBeenCalledOnce();
});
