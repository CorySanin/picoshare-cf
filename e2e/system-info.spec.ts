import { test, expect } from "./fixtures";
import { login } from "./helpers/login";

test("views system info", async ({ page }) => {
  await login(page);

  await page.getByRole("menuitem", { name: "System" }).hover();
  await page.getByRole("menuitem", { name: "Information" }).click();
  await expect(page).toHaveURL("/information");
  await expect(
    page.getByRole("heading", { name: "System Information", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("progressbar", { name: "Disk usage" }),
  ).toBeVisible();
});
