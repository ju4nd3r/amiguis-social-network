import { test, expect } from "@playwright/test";

test.describe("Amiguis E2E Critical User Journeys", () => {
  test("renders Pinky Promise logo, hides technical jargon, and toggles dark/light mode", async ({
    page,
  }) => {
    await page.goto("/");

    // Minimalist Pinky Promise logo is visible
    await expect(
      page.locator('svg[aria-label="Amiguis Pinky Promise Logo"]')
    ).toBeVisible();
    await expect(page.locator("header")).toContainText("Amiguis");

    // Technical text and old badges must NOT be shown
    await expect(page.getByText("100% Cronológico")).toHaveCount(0);
    await expect(page.getByText(/created_at DESC/i)).toHaveCount(0);
    await expect(
      page.getByText(
        "Únete para publicar, seguir parceros y armar tus círculos."
      )
    ).toHaveCount(0);

    // Toggle Dark / Light mode
    const html = page.locator("html");
    await expect(html).toHaveClass(/dark/);
    await page.getByTestId("theme-toggle").click();
    await expect(html).not.toHaveClass(/dark/);
    await page.getByTestId("theme-toggle").click();
    await expect(html).toHaveClass(/dark/);
  });

  test("only allows 'Interesante' reaction and opens/submits comments without errors", async ({
    page,
  }) => {
    const pageErrors: Error[] = [];
    page.on("pageerror", (err) => pageErrors.push(err));

    await page.goto("/");

    // Verify only "Interesante" reaction exists (no Me gusta, Parce, Arepa)
    await expect(page.getByRole("button", { name: /Interesante/i }).first()).toBeVisible();
    await expect(page.getByRole("button", { name: /Me gusta/i })).toHaveCount(0);
    await expect(page.getByRole("button", { name: /Arepa/i })).toHaveCount(0);

    // Click "Interesante" reaction
    const firstReactBtn = page.getByRole("button", { name: /Interesante/i }).first();
    await firstReactBtn.click();

    // Click the comment button on the first post ("Comentarios (2)") and ensure the 2 existing comments are shown
    const firstCommentToggle = page
      .getByRole("button", { name: /Comentarios \(2\)/i })
      .first();
    await expect(firstCommentToggle).toBeVisible();
    await firstCommentToggle.click();

    await expect(page.getByText(/No hay comentarios aún/i)).toHaveCount(0);
    await expect(
      page.getByText(/Totalmente de acuerdo, Cami/i)
    ).toBeVisible();
    await expect(
      page.getByText(/Un abrazo desde Cali, amigui/i)
    ).toBeVisible();

    // Submit a new comment
    const commentInput = page
      .getByPlaceholder("Escribe un comentario...")
      .first();
    await expect(commentInput).toBeVisible();
    await commentInput.fill("¡Excelente publicación, amigui!");
    await page.getByRole("button", { name: "Enviar" }).first().click();

    // Verify the new comment is rendered in the post's comment list and count increments to 3
    await expect(
      page.getByText("¡Excelente publicación, amigui!")
    ).toBeVisible();
    await expect(
      page.getByRole("button", { name: /Comentarios \(3\)/i }).first()
    ).toBeVisible();
    expect(pageErrors).toHaveLength(0);
  });

  test("provides drag-and-drop zone and PC file selector button for photos and videos", async ({
    page,
  }) => {
    await page.goto("/");

    // Switch post composer to Photo mode
    await page.getByRole("button", { name: "Foto", exact: true }).click();
    await expect(page.getByTestId("media-dropzone")).toBeVisible();
    await expect(page.getByTestId("select-file-button")).toContainText(
      "Seleccionar foto desde tu PC"
    );

    // Switch post composer to Video mode
    await page.getByRole("button", { name: "Video", exact: true }).click();
    await expect(page.getByTestId("media-dropzone")).toBeVisible();
    await expect(page.getByTestId("select-file-button")).toContainText(
      "Seleccionar video desde tu PC"
    );
  });

  test("register new amigui account flow", async ({ page }) => {
    await page.goto("/login");
    await page.getByRole("button", { name: "Registrarse" }).click();

    const uniqueUser = `amigui_${Date.now()}`;
    await page
      .getByPlaceholder("camila@ejemplo.co")
      .fill(`${uniqueUser}@amiguis.co`);
    await page.getByPlaceholder("Camila Restrepo").fill("Amigui Colombiano");
    await page.getByPlaceholder("camila_medallo").fill(uniqueUser);
    await page.getByPlaceholder("••••••••").fill("ClaveSegura123!");
    await page.getByRole("checkbox").check();
    await page.getByRole("button", { name: "Crear mi cuenta" }).click();
  });
});
