/**
 * Unit tests for getWebviewContent.
 *
 * Test spec: spec/test/vscode/src/views/getWebviewContent.md
 * Source under test: vscode/src/views/getWebviewContent.ts
 *
 * Scaffolded: The source file does not yet exist. These tests are structured
 * to compile and provide coverage once the production surface is created.
 *
 * Missing production surface:
 *   - vscode/src/views/getWebviewContent.ts
 *   - getWebviewContent function
 */
import * as sinon from "sinon";
import { expect } from "chai";

import {
  createStubWebview,
  createStubExtensionUri,
  extractNonceFromCsp,
  extractNonceFromStyleTag,
  extractNonceFromScriptTag,
  EXPECTED_ELEMENT_IDS,
  type StubWebview,
  type StubUri,
} from "./helpers/webviewStubs";

import { getWebviewContent } from "../../src/views/getWebviewContent";

describe("getWebviewContent", function () {
  let stubWebview: StubWebview;
  let stubExtensionUri: StubUri;

  beforeEach(function () {
    stubWebview = createStubWebview("https://test.csp");
    stubExtensionUri = createStubExtensionUri();
  });

  // ─── Helper: invoke the function under test ──────────────────────────────
  function invoke(
    webview: StubWebview = stubWebview,
    uri: StubUri = stubExtensionUri,
  ): string {
    return getWebviewContent(webview as any, uri as any);
  }

  // ─── Happy Path — getWebviewContent ──────────────────────────────────────

  describe("Happy Path — getWebviewContent", function () {
    it("should return a valid HTML5 document", function () {
      const html = invoke();

      expect(html).to.match(/^<!DOCTYPE html>/);
      expect(html).to.contain("<html");
      expect(html).to.contain("<head");
      expect(html).to.contain("<body");
    });

    it("should include CSP meta tag with nonce-gated style-src and script-src", function () {
      const html = invoke();

      expect(html).to.contain("<meta");
      expect(html).to.contain("default-src 'none'");
      expect(html).to.match(/style-src\s[^;]*'nonce-/);
      expect(html).to.match(/script-src\s[^;]*'nonce-/);
    });

    it("should include a style block with matching nonce", function () {
      const html = invoke();

      const cspNonce = extractNonceFromCsp(html);
      const styleNonce = extractNonceFromStyleTag(html);

      expect(cspNonce).to.not.be.null;
      expect(styleNonce).to.not.be.null;
      expect(styleNonce).to.equal(cspNonce);

      // Exactly one style tag with nonce
      const styleMatches = html.match(/<style\s+nonce="/g);
      expect(styleMatches).to.have.lengthOf(1);
    });

    it("should include a script block with matching nonce", function () {
      const html = invoke();

      const cspNonce = extractNonceFromCsp(html);
      const scriptNonce = extractNonceFromScriptTag(html);

      expect(cspNonce).to.not.be.null;
      expect(scriptNonce).to.not.be.null;
      expect(scriptNonce).to.equal(cspNonce);

      // Exactly one script tag with nonce
      const scriptMatches = html.match(/<script\s+nonce="/g);
      expect(scriptMatches).to.have.lengthOf(1);
    });

    it("should generate a unique nonce per invocation", function () {
      const html1 = invoke();
      const html2 = invoke();

      const nonce1 = extractNonceFromCsp(html1);
      const nonce2 = extractNonceFromCsp(html2);

      expect(nonce1).to.not.be.null;
      expect(nonce2).to.not.be.null;
      expect(nonce1).to.not.equal(nonce2);
    });

    it("should contain header element with text Spectra", function () {
      const html = invoke();

      // The header element should contain the text "Spectra"
      expect(html).to.match(
        /<h[1-6][^>]*>.*Spectra.*<\/h[1-6]>|<header[^>]*>.*Spectra.*<\/header>/s,
      );
    });

    it("should contain sessions list page element", function () {
      const html = invoke();
      expect(html).to.match(/id=["']page-sessions["']/);
    });

    it("should contain session detail page element", function () {
      const html = invoke();
      expect(html).to.match(/id=["']page-detail["']/);
    });

    it("should contain workflow-select dropdown", function () {
      const html = invoke();
      expect(html).to.match(/<select[^>]*id=["']workflow-select["']/);
    });

    it("should contain Run button", function () {
      const html = invoke();
      expect(html).to.match(/id=["']btn-run["']/);
    });

    it("should contain session-list container", function () {
      const html = invoke();
      expect(html).to.match(/id=["']session-list["']/);
    });

    it("should contain back button on detail page", function () {
      const html = invoke();
      expect(html).to.match(/id=["']btn-back["']/);
    });

    it("should contain event-list container on detail page", function () {
      const html = invoke();
      expect(html).to.match(/id=["']event-list["']/);
    });

    it("should contain event-type-select dropdown on detail page", function () {
      const html = invoke();
      expect(html).to.match(/<select[^>]*id=["']event-type-select["']/);
    });

    it("should contain event-message-input on detail page", function () {
      const html = invoke();
      expect(html).to.match(/<input[^>]*id=["']event-message-input["']/);
    });

    it("should contain Send button on detail page", function () {
      const html = invoke();
      expect(html).to.match(/id=["']btn-send["']/);
    });

    it("should include acquireVsCodeApi call in script", function () {
      const html = invoke();
      expect(html).to.contain("acquireVsCodeApi()");
    });

    it("should include message event listener in script", function () {
      const html = invoke();

      const hasMessageListener =
        html.includes("addEventListener('message'") ||
        html.includes('addEventListener("message"');

      expect(hasMessageListener).to.be.true;
    });

    it("should not contain inline event handlers", function () {
      const html = invoke();

      expect(html).to.not.contain("onclick=");
      expect(html).to.not.contain("onsubmit=");
      expect(html).to.not.contain("onchange=");
      expect(html).to.not.contain("onkeydown=");
    });

    it("should not use eval in script", function () {
      const html = invoke();
      expect(html).to.not.contain("eval(");
    });
  });

  // ─── Mock / Dependency Interaction ──────────────────────────────────────────

  describe("Mock / Dependency Interaction", function () {
    it("should access webview.cspSource", function () {
      // Use a property spy to detect access
      const webview: any = {};
      let cspAccessed = false;
      Object.defineProperty(webview, "cspSource", {
        get() {
          cspAccessed = true;
          return "https://csp.test";
        },
        enumerable: true,
        configurable: true,
      });

      getWebviewContent(webview, stubExtensionUri as any);

      expect(cspAccessed).to.be.true;
    });
  });

  // ─── Null / Empty Input ─────────────────────────────────────────────────────

  describe("Null / Empty Input", function () {
    it("should produce valid HTML when cspSource is empty string", function () {
      const emptyWebview = createStubWebview("");
      const html = getWebviewContent(
        emptyWebview as any,
        stubExtensionUri as any,
      );

      expect(html).to.match(/^<!DOCTYPE html>/);
      expect(html).to.contain("<meta");
      // Nonce-based parts should still be present
      expect(html).to.match(/nonce-[a-f0-9]+/);
    });
  });

  // ─── Idempotency ───────────────────────────────────────────────────────────

  describe("Idempotency", function () {
    it("should produce structurally consistent output on repeated calls", function () {
      const html1 = invoke();
      const html2 = invoke();

      // Both results should contain the same set of element IDs
      for (const id of EXPECTED_ELEMENT_IDS) {
        const re = new RegExp(`id=["']${id}["']`);
        expect(html1).to.match(re, `First call missing id="${id}"`);
        expect(html2).to.match(re, `Second call missing id="${id}"`);
      }
    });
  });
});
