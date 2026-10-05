import van from "vanjs-core";

const { a, button, code, div, footer, h1, header, label, p, pre, section, span, textarea } = van.tags;

type Detect = { ratio: number; verdict: string; decoded: string };
type LeetAPI = {
  encode(text: string, level: string, seed: number): string;
  decode(text: string): string;
  detect(text: string): Detect;
  frame(text: string, progress: number, seed: number): string;
  durationMs: number;
};
declare const Go: new () => { importObject: WebAssembly.Imports; run(i: WebAssembly.Instance): Promise<void> };
declare global {
  interface Window { leet?: LeetAPI }
}

const LEVELS = ["basic", "advanced", "elite"] as const;
const app = document.getElementById("app")!;

const ready = van.state(false);
const failed = van.state("");
const input = van.state("Hack the planet");
const decoding = van.state(false);
const level = van.state<string>("elite");
const random = van.state(false);
const seed = van.state(1 + Math.floor(Math.random() * 2 ** 31));
const copied = van.state(false);
const shared = van.state("");
const progress = van.state(1); // decrypting animation: 1 = idle
let animSeed = 0;
let animFrame = 0;

const output = van.derive(() => {
  if (!ready.val) return failed.val || "loading WebAssembly...";
  const api = window.leet!;
  return decoding.val ? api.decode(input.val) : api.encode(input.val, level.val, random.val ? seed.val : 0);
});
const score = van.derive(() => (ready.val ? window.leet!.detect(input.val) : { ratio: 0, verdict: "", decoded: "" }));

const reducedMotion = matchMedia("(prefers-reduced-motion: reduce)");

// Replays the decrypting effect on the output panel.
const animate = () => {
  if (reducedMotion.matches) return;
  cancelAnimationFrame(animFrame);
  animSeed = (animSeed + 1) % 2 ** 31;
  const duration = window.leet?.durationMs ?? 2000;
  const start = performance.now();
  const step = (now: number) => {
    progress.val = Math.min((now - start) / duration, 1);
    if (progress.val < 1) animFrame = requestAnimationFrame(step);
  };
  animFrame = requestAnimationFrame(step);
};

// Wraps a control's handler so every change replays the animation.
const act = (fn: () => void) => () => {
  fn();
  animate();
};

const pill = (text: string | (() => string), on: () => boolean, onclick: () => void, disabled: () => boolean = () => false) =>
  button({
    class: () =>
      "rounded-full px-3 py-1 text-sm transition " +
      (disabled() ? "cursor-not-allowed opacity-30 " : "") +
      (on() ? "bg-leet text-neutral-950 font-bold" : "text-neutral-400 hover:text-leet"),
    disabled,
    onclick,
  }, text);

const group = (...children: HTMLElement[]) =>
  div({ class: "flex flex-wrap items-center gap-1 rounded-full border border-neutral-800 p-1" }, ...children);

const panelClass = "rounded-xl border bg-neutral-900/60 p-4 min-h-48 flex flex-col gap-2";
const titleClass = "text-xs font-bold tracking-widest text-leet";

// Created once, outside any derive, so typing never recreates it (keeps focus).
const editor = textarea({
  class: "flex-1 resize-none bg-transparent text-lg outline-none placeholder:text-neutral-600",
  name: "text",
  placeholder: "Type something...",
  spellcheck: false,
  value: input,
  oninput: (e: Event) => (input.val = (e.target as HTMLTextAreaElement).value),
});

const swap = act(() => {
  input.val = output.rawVal;
  decoding.val = !decoding.rawVal;
});

// Share links carry the 1337, never the plain text, and open in decode mode:
// whoever opens one watches the message decrypt.
const shareLink = () => {
  const leetText = decoding.rawVal ? input.rawVal : output.rawVal;
  return `${location.origin}${location.pathname}#${new URLSearchParams({ d: leetText })}`;
};

const flash = (msg: string) => {
  shared.val = msg;
  setTimeout(() => (shared.val = ""), 1800);
};

const share = async () => {
  const url = shareLink();
  if (navigator.share && matchMedia("(pointer: coarse)").matches) {
    try {
      await navigator.share({ title: "A 1337 message for you", url });
    } catch {
      // dismissed by the user
    }
    return;
  }
  await navigator.clipboard.writeText(url);
  flash("link copied!");
};

const loadFromHash = () => {
  const d = new URLSearchParams(location.hash.slice(1)).get("d");
  if (d === null) return false;
  input.val = d;
  decoding.val = true;
  return true;
};

const copy = async () => {
  await navigator.clipboard.writeText(output.rawVal);
  copied.val = true;
  setTimeout(() => (copied.val = false), 1500);
};

van.add(app,
  header({ class: "mb-8 flex flex-wrap items-end justify-between gap-4" },
    div(
      h1({ class: "text-5xl font-bold text-leet" }, "1337"),
      p({ class: "mt-2 text-neutral-400" }, "Text to leet speak and back. Go compiled to WebAssembly, nothing leaves your browser."),
    ),
    a({ class: "text-sm text-neutral-400 underline hover:text-leet", href: app.dataset.repo! }, "source on GitHub"),
  ),

  section({ class: "mb-4 flex flex-wrap items-center gap-3" },
    group(
      pill("encode", () => !decoding.val, act(() => (decoding.val = false))),
      pill("decode", () => decoding.val, act(() => (decoding.val = true))),
    ),
    group(...LEVELS.map(l => pill(l, () => level.val === l, act(() => (level.val = l)), () => decoding.val))),
    group(
      pill("random", () => random.val, act(() => (random.val = !random.val)), () => decoding.val),
      pill("reshuffle", () => false, act(() => (seed.val = seed.rawVal + 1)), () => decoding.val || !random.val),
    ),
  ),

  div({ class: "grid gap-4 md:grid-cols-2" },
    label({ class: panelClass + " border-leet/60" },
      span({ class: titleClass }, () => (decoding.val ? "1337" : "PLAIN")),
      editor,
    ),
    div({ class: panelClass + " border-neutral-800" },
      div({ class: "flex items-center justify-between" },
        span({ class: titleClass }, () => (decoding.val ? "PLAIN" : "1337")),
        div({ class: "flex gap-1" },
          pill("swap", () => false, swap),
          pill(() => (copied.val ? "copied!" : "copy"), () => copied.val, copy),
          pill(() => shared.val || "share", () => shared.val !== "", share),
        ),
      ),
      pre({ class: "flex-1 whitespace-pre-wrap break-all text-lg text-leet" },
        () => (progress.val < 1 && ready.val ? window.leet!.frame(output.val, progress.val, animSeed) : output.val)),
    ),
  ),

  section({ class: "mt-6" },
    div({ class: "mb-1 flex justify-between text-xs text-neutral-400" },
      span("1337-o-meter (input)"),
      span(() => `${Math.round(score.val.ratio * 100)}% · ${score.val.verdict}`),
    ),
    div({ class: "h-2 overflow-hidden rounded-full bg-neutral-800" },
      div({ class: "h-full bg-leet transition-all", style: () => `width:${score.val.ratio * 100}%` }),
    ),
  ),

  footer({ class: "mt-12 space-y-2 text-sm text-neutral-400" },
    p("Prefer the terminal? Same engine, plus a live TUI:"),
    pre({ class: "overflow-x-auto rounded-lg bg-neutral-900 p-3 text-neutral-200" },
      code("go install github.com/carlosprados/go-1337/cmd/leet@latest\nleet            # live editor\nleet encode -l basic --random \"leet speak\"\nleet share \"meet me at the usual place\"   # prints a link like the share button"),
    ),
  ),
);

loadFromHash();
addEventListener("hashchange", () => loadFromHash() && animate());

const go = new Go();
WebAssembly.instantiateStreaming(fetch(app.dataset.wasm!), go.importObject)
  .then(r => {
    addEventListener("leet-ready", () => {
      ready.val = true;
      animate();
    }, { once: true });
    void go.run(r.instance);
  })
  .catch(err => (failed.val = `could not load WebAssembly: ${err}`));
