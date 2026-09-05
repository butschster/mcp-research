# Dovod application identity

The approved wordmark is the uncut, unmirrored Bricolage Grotesque 800 study from `docs/brand/dovod/type/01-ink.svg`. `components/BrandLogo.vue` uses its exact outlines and inherits the surrounding text color.

- `dovod.svg` / `dovod-inverse.svg`: forest and paper versions.
- `icon.svg` / PNG sizes: the same lowercase d, used alone for browser and installation icons.
- `social-card.svg` / `.png`: share preview.
- `fonts/`: local Bricolage Grotesque display face and its SIL OFL license. Wordmark SVGs require no installed font.

Forest `#243D32`, paper `#F6F3EC`, sage `#C7D9A7`. All SPA components share the theme roles in `assets/css/tokens.css`. Light uses paper and forest ink. Dark uses near-black `#0D1117`, neutral panels `#151B23`, and pale text `#F0F6FC`, with sage reserved for accents. The neutral ramp is informed by [GitHub Primer](https://github.com/primer/primitives/blob/main/src/tokens/base/color/dark/dark.json5).

`data-theme` on the root selects the palette. `useTheme` persists the choice under `dovod-theme`; `ThemeToggle` is shared by navigation and auth. Canvas and Mermaid resolve the same roles through `utils/theme.ts`. Storybook exposes both themes in its toolbar. The standalone OAuth template carries a small matching fallback palette because it must work without the SPA bundle.
