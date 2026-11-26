import { test, expect } from "./fixtures";

test("serves icons and the web manifest", async ({ page, request }) => {
  await page.goto("/");

  {
    const response = await request.get("/favicon.ico");
    expect(response.status()).toBe(200);
  }

  {
    const response = await request.get("/android-chrome-192x192.png");
    expect(response.status()).toBe(200);
  }

  {
    const response = await request.get("/apple-touch-icon.png");
    expect(response.status()).toBe(200);
  }

  {
    const response = await request.get("/site.webmanifest");
    expect(response.status()).toBe(200);
  }
});
