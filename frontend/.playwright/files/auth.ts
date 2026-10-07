import { expect, type APIRequestContext, type APIResponse } from '@playwright/test';

// The disposable server retains its production login rate limit. Retry only its
// explicit 429 response when a full desktop/mobile suite exhausts the burst.
export async function loginForTest(request: APIRequestContext, cookie = false): Promise<APIResponse> {
  let response: APIResponse | undefined;
  await expect.poll(async () => {
    response = await request.post('/api/auth/login', {headers:cookie?{'X-Edda-Session':'cookie'}:undefined,data:{email:'browser@example.invalid',password:'browser-test-password'}});
    if (response.status() === 429) return false;
    expect(response.ok()).toBeTruthy();
    return true;
  }, {intervals:[2100], timeout:10000}).toBe(true);
  if (!response) throw new Error('Missing login response');
  return response;
}
