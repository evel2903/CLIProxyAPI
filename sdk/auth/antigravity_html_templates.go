package auth

// antigravityLoginSuccessHtml is the HTML page served to the browser after a
// successful Antigravity OAuth callback. It mirrors the styling used by the
// Codex and Claude login pages so the three flows feel like one product.
const antigravityLoginSuccessHtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Antigravity Connected</title>
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%232f6b3f'%3E%3Cpath d='M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z'/%3E%3C/svg%3E">
    <style>
        :root {
            --bg: #faf9f5;
            --surface: #fffdf8;
            --ink: #2d2a26;
            --ink-soft: #7a746c;
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
        }
        @keyframes rise {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
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
            transition: opacity 0.12s ease;
        }
        .button:focus-visible { outline: 2px solid var(--ok); outline-offset: 2px; }
        .button-primary { background: var(--accent); color: var(--accent-fg); }
        .button-primary:hover { opacity: 0.88; }
        .footer {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-top: 1.35rem;
            padding-top: 1rem;
            border-top: 1px solid var(--line);
            font-family: var(--font-mono);
            font-size: 0.7rem;
            color: var(--ink-soft);
        }
        .footer #countdown { font-variant-numeric: tabular-nums; }
    </style>
</head>
<body>
    <main class="card" role="status" aria-live="polite">
        <div class="status-line">
            <span class="dot" aria-hidden="true"></span>
            <span>evel auth antigravity&nbsp;</span><span class="ok">— connected</span>
        </div>

        <div class="head">
            <span class="mark" aria-hidden="true">
                <svg viewBox="0 0 24 24"><path d="M5 12.5l4.5 4.5L19 7"/></svg>
            </span>
            <div>
                <h1>Antigravity is signed in</h1>
                <p class="sub">This browser tab did its job — your Google account is now linked to EvelProxyTool via Antigravity. You can close it and get back to your editor.</p>
            </div>
        </div>

        <div class="actions">
            <button class="button button-primary" type="button" onclick="attemptClose()">Close this tab</button>
        </div>

        <div class="footer">
            <span>EvelProxyTool</span>
            <span id="foot-status">closing in <span id="countdown">10</span>s</span>
        </div>
    </main>

    <script>
        var seconds = 10;
        var countdownEl = document.getElementById('countdown');
        var status = document.getElementById('foot-status');
        var timer = null;

        function attemptClose() {
            clearInterval(timer);
            window.close();
            setTimeout(function () {
                status.textContent = 'you can close this tab now';
            }, 150);
        }

        timer = setInterval(function () {
            seconds -= 1;
            countdownEl.textContent = seconds;
            if (seconds <= 0) attemptClose();
        }, 1000);

        document.addEventListener('keydown', function (e) {
            if (e.key === 'Escape') attemptClose();
        });

        document.querySelector('.button-primary').focus();
    </script>
</body>
</html>`

// antigravityLoginFailureHtml is served when the Antigravity OAuth callback
// arrives without a usable authorization code.
const antigravityLoginFailureHtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Antigravity Sign-In Failed</title>
    <style>
        :root {
            --bg: #faf9f5;
            --surface: #fffdf8;
            --ink: #2d2a26;
            --ink-soft: #7a746c;
            --line: #e3e1db;
            --bad: #9f2d20;
            --bad-soft: #fbecea;
            --bad-line: #e3b3ab;
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
                --line: #313033;
                --bad: #f87171;
                --bad-soft: rgba(248, 113, 113, 0.12);
                --bad-line: rgba(248, 113, 113, 0.32);
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
        .status-line .bad { color: var(--bad); font-weight: 500; }
        .head { display: flex; align-items: flex-start; gap: 0.9rem; }
        .mark {
            flex: none;
            width: 2.5rem;
            height: 2.5rem;
            border-radius: 50%;
            background: var(--bad-soft);
            border: 1px solid var(--bad-line);
            display: grid;
            place-items: center;
            font-size: 1.1rem;
            color: var(--bad);
        }
        h1 { font-size: 1.28rem; font-weight: 700; letter-spacing: -0.01em; line-height: 1.3; margin: 0.35rem 0 0.4rem; }
        .sub { margin: 0; color: var(--ink-soft); font-size: 0.92rem; line-height: 1.55; }
        .footer {
            margin-top: 1.35rem;
            padding-top: 1rem;
            border-top: 1px solid var(--line);
            font-family: var(--font-mono);
            font-size: 0.7rem;
            color: var(--ink-soft);
        }
    </style>
</head>
<body>
    <main class="card" role="alert">
        <div class="status-line">
            <span>evel auth antigravity&nbsp;</span><span class="bad">— failed</span>
        </div>
        <div class="head">
            <span class="mark" aria-hidden="true">!</span>
            <div>
                <h1>Sign-in didn't go through</h1>
                <p class="sub">Antigravity didn't return an authorization code. Check the EvelProxyTool window for details and try signing in again.</p>
            </div>
        </div>
        <div class="footer">EvelProxyTool</div>
    </main>
</body>
</html>`
