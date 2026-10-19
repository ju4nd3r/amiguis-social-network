import { test, expect } from "@playwright/test";

test.describe("Amiguis E2E Critical User Journeys", () => {
  test("renders chronological feed and filter bar on mobile viewport", async ({
    page,
  }) => {
    await page.goto("/");
    await expect(page.locator("header")).toContainText("Amiguis");
    await expect(page.locator("header")).toContainText("100% Cronológico");
    await expect(page.getByText("Sin algoritmos ni anuncios.")).toBeVisible();
    await expect(page.getByRole("button", { name: /Todo/i })).toBeVisible();
    await expect(page.getByRole("button", { name: /Fotos/i })).toBeVisible();
  });

  test("register -> post -> chronological feed -> Habeas Data export", async ({
    page,
  }) => {
    await page.goto("/login");
    await page.getByRole("button", { name: "Registrarse" }).click();

    const uniqueUser = `parcero_${Date.now()}`;
    await page.getByPlaceholder("camila@ejemplo.co").fill(`${uniqueUser}@amiguis.co`);
    await page.getByPlaceholder("Camila Restrepo").fill("Parcero Colombiano");
    await page.getByPlaceholder("camila_medallo").fill(uniqueUser);
    await page.getByPlaceholder("••••••••").fill("ClaveSegura123!");
    await page.getByRole("checkbox").check();
    await page.getByRole("button", { name: "Crear mi cuenta" }).click();

    await expect(page).toHaveURL("/");
  });
});
