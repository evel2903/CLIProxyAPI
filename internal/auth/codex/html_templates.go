package codex

// LoginSuccessHTML is the HTML template for the page shown after a successful
// OAuth2 authentication with Codex. It informs the user that the authentication
// was successful and provides a countdown timer to automatically close the window.
const LoginSuccessHtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Codex Connected</title>
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%232f6b3f'%3E%3Cpath d='M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z'/%3E%3C/svg%3E">
    <style>
        :root {
            --bg: #faf9f5;
            --surface: #fffdf8;
            --ink: #2d2a26;
            --ink-soft: #7a746c;
            --ink-faint: #a49e93;
            --line: #e3e1db;
            --accent: #18181b;
            --accent-fg: #fafafa;
            --ok: #2f6b3f;
            --ok-soft: #f1f8f1;
            --ok-line: #b8d1bb;
            --shadow: 0 1px 2px rgba(45, 42, 38, 0.04), 0 12px 32px -16px rgba(45, 42, 38, 0.18);
            --font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            --font-mono: ui-monospace, SFMono-Regular, "Cascadia Code", "Segoe UI Mono", Consolas, "Liberation Mono", monospace;
        }
        @media (prefers-color-scheme: dark) {
            :root {
                --bg: #18181b;
                --surface: #201f22;
                --ink: #f4f4f5;
                --ink-soft: #a3a19c;
                --ink-faint: #6c6a67;
                --line: #313033;
                --accent: #fafafa;
                --accent-fg: #18181b;
                --ok: #4ade80;
                --ok-soft: rgba(74, 222, 128, 0.12);
                --ok-line: rgba(74, 222, 128, 0.32);
                --shadow: 0 1px 2px rgba(0, 0, 0, 0.3), 0 20px 44px -20px rgba(0, 0, 0, 0.55);
            }
        }
        * { box-sizing: border-box; }
        html, body { margin: 0; height: 100%; }
        body {
            background: var(--bg);
            color: var(--ink);
            font-family: var(--font-sans);
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 1.5rem;
            -webkit-font-smoothing: antialiased;
        }
        @media (prefers-reduced-motion: no-preference) {
            .card { animation: rise 0.42s cubic-bezier(0.16, 1, 0.3, 1) both; }
            .mark path { animation: draw 0.5s 0.15s cubic-bezier(0.65, 0, 0.35, 1) both; }
        }
        @keyframes rise {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }
        @keyframes draw {
            from { stroke-dashoffset: 24; }
            to { stroke-dashoffset: 0; }
        }
        .card {
            width: 100%;
            max-width: 26rem;
            background: var(--surface);
            border: 1px solid var(--line);
            border-radius: 14px;
            box-shadow: var(--shadow);
            padding: 1.75rem 1.75rem 1.5rem;
        }
        .status-line {
            display: flex;
            align-items: center;
            gap: 0.55rem;
            font-family: var(--font-mono);
            font-size: 0.72rem;
            letter-spacing: 0.02em;
            color: var(--ink-soft);
            padding-bottom: 1.1rem;
            margin-bottom: 1.15rem;
            border-bottom: 1px solid var(--line);
        }
        .status-line .ok { color: var(--ok); font-weight: 500; }
        .dot {
            width: 6px;
            height: 6px;
            border-radius: 50%;
            background: var(--ok);
            flex: none;
            box-shadow: 0 0 0 3px var(--ok-soft);
        }
        .head {
            display: flex;
            align-items: flex-start;
            gap: 0.9rem;
            margin-bottom: 1.35rem;
        }
        .mark {
            flex: none;
            width: 2.5rem;
            height: 2.5rem;
            border-radius: 50%;
            background: var(--ok-soft);
            border: 1px solid var(--ok-line);
            display: grid;
            place-items: center;
        }
        .mark svg { width: 1.15rem; height: 1.15rem; }
        .mark path {
            fill: none;
            stroke: var(--ok);
            stroke-width: 2.4;
            stroke-linecap: round;
            stroke-linejoin: round;
            stroke-dasharray: 24;
        }
        h1 {
            font-size: 1.28rem;
            font-weight: 700;
            letter-spacing: -0.01em;
            line-height: 1.3;
            margin: 0.35rem 0 0.4rem;
        }
        .sub {
            margin: 0;
            color: var(--ink-soft);
            font-size: 0.92rem;
            line-height: 1.55;
        }
        .setup-notice {
            display: flex;
            gap: 0.65rem;
            background: var(--ok-soft);
            border: 1px solid var(--ok-line);
            border-radius: 10px;
            padding: 0.85rem 0.95rem;
            margin-bottom: 1.35rem;
        }
        .setup-notice svg { flex: none; width: 1rem; height: 1rem; margin-top: 0.15rem; stroke: var(--ok); }
        .setup-notice p { margin: 0; font-size: 0.82rem; line-height: 1.5; color: var(--ink); }
        .setup-notice a { color: inherit; text-decoration: underline; text-underline-offset: 2px; }
        .actions { display: flex; flex-direction: column; gap: 0.55rem; }
        .button {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 0.5rem;
            width: 100%;
            border-radius: 8px;
            padding: 0.7rem 1rem;
            font-family: inherit;
            font-size: 0.86rem;
            font-weight: 600;
            text-decoration: none;
            cursor: pointer;
            border: 1px solid transparent;
            transition: transform 0.12s ease, background 0.12s ease, border-color 0.12s ease, opacity 0.12s ease;
        }
        .button:active { transform: translateY(1px); }
        .button:focus-visible { outline: 2px solid var(--ok); outline-offset: 2px; }
        .button-primary { background: var(--accent); color: var(--accent-fg); }
        .button-primary:hover { opacity: 0.88; }
        .button-secondary { background: transparent; color: var(--ink-soft); border-color: var(--line); }
        .button-secondary:hover { color: var(--ink); border-color: var(--ink-faint); }
        .footer {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-top: 1.35rem;
            padding-top: 1rem;
            border-top: 1px solid var(--line);
            font-family: var(--font-mono);
            font-size: 0.7rem;
            color: var(--ink-faint);
        }
        .footer #countdown { font-variant-numeric: tabular-nums; color: var(--ink-soft); }
        .footer a { color: inherit; }
    </style>
</head>
<body>
    <main class="card" role="status" aria-live="polite">
        <div class="status-line">
            <span class="dot" aria-hidden="true"></span>
            <span>evel auth codex&nbsp;</span><span class="ok">— connected</span>
        </div>

        <div class="head">
            <span class="mark" aria-hidden="true">
                <svg viewBox="0 0 24 24"><path d="M5 12.5l4.5 4.5L19 7"/></svg>
            </span>
            <div>
                <h1>Codex is signed in</h1>
                <p class="sub">This browser tab did its job — your Codex account is now linked to EvelProxyTool. You can close it and get back to your editor.</p>
            </div>
        </div>

        {{SETUP_NOTICE}}

        <div class="actions">
            <button class="button button-primary" type="button" onclick="attemptClose()">Close this tab</button>
            <a href="{{PLATFORM_URL}}" target="_blank" rel="noopener" class="button button-secondary">Open OpenAI platform ↗</a>
        </div>

        <div class="footer">
            <span>EvelProxyTool</span>
            <span id="foot-status">closing in <span id="countdown">10</span>s — <a href="#" id="cancel-close">stay</a></span>
        </div>
    </main>

    <script>
        var seconds = 10;
        var countdownEl = document.getElementById('countdown');
        var status = document.getElementById('foot-status');
        var cancelLink = document.getElementById('cancel-close');
        var timer = null;

        function attemptClose() {
            clearInterval(timer);
            window.close();
            // Browsers ignore close() on tabs they didn't open via script — assume that
            // happened if we're still here a beat later, and stop the fake countdown.
            setTimeout(function () {
                status.textContent = 'you can close this tab now';
            }, 150);
        }

        timer = setInterval(function () {
            seconds -= 1;
            countdownEl.textContent = seconds;
            if (seconds <= 0) attemptClose();
        }, 1000);

        cancelLink.addEventListener('click', function (e) {
            e.preventDefault();
            clearInterval(timer);
            status.textContent = 'staying open';
        });

        document.addEventListener('keydown', function (e) {
            if (e.key === 'Escape') attemptClose();
        });

        document.querySelector('.button-primary').focus();
    </script>
</body>
</html>`

// SetupNoticeHTML is the HTML template for the section that provides instructions
// for additional setup. This is displayed on the success page when further actions
// are required from the user.
const SetupNoticeHtml = `
        <div class="setup-notice">
            <svg viewBox="0 0 24 24" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9"/><path d="M12 8v5M12 16h.01"/></svg>
            <p>First time with this account? Finish setup on <a href="{{PLATFORM_URL}}" target="_blank" rel="noopener">Codex</a> before it can serve requests.</p>
        </div>`
