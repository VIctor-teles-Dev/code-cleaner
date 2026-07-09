import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { ContactForm } from "./contact-form";

async function fillAndSubmit() {
  const user = userEvent.setup();
  await user.type(screen.getByLabelText("Nome"), "Ana");
  await user.type(screen.getByLabelText("Email"), "ana@example.com");
  await user.type(screen.getByLabelText("Mensagem"), "Olá, Victor!");
  await user.click(screen.getByRole("button", { name: /enviar mensagem/i }));
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("ContactForm", () => {
  it("submits the message and shows the success state", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ status: "received" }), { status: 201 }),
    );
    vi.stubGlobal("fetch", fetchMock);

    render(<ContactForm />);
    await fillAndSubmit();

    expect(await screen.findByText(/mensagem enviada/i)).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/contact",
      expect.objectContaining({ method: "POST" }),
    );
    const payload = JSON.parse(fetchMock.mock.calls[0]?.[1]?.body as string);
    expect(payload).toEqual({
      name: "Ana",
      email: "ana@example.com",
      message: "Olá, Victor!",
    });
  });

  it("shows the backend validation message on error", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({ status: "invalid", error: "email inválido" }),
          { status: 400 },
        ),
      ),
    );

    render(<ContactForm />);
    await fillAndSubmit();

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "email inválido",
    );
  });

  it("shows a generic error when the request fails", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new Error("network down")),
    );

    render(<ContactForm />);
    await fillAndSubmit();

    expect(await screen.findByRole("alert")).toHaveTextContent(
      /não foi possível enviar/i,
    );
  });
});
