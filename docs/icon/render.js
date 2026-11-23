// Renders assets/icon.svg to PNGs with Chromium (Playwright).
// Usage: node docs/icon/render.js assets/icon.svg <out-dir>
const { chromium } = require('playwright');
const fs = require('fs');
(async () => {
  const [src, outDir] = process.argv.slice(2);
  const svg = fs.readFileSync(src, 'utf8');
  const browser = await chromium.launch();
  const page = await browser.newPage({ deviceScaleFactor: 1 });
  for (const size of [1024, 512, 256, 64, 32]) {
    await page.setViewportSize({ width: size, height: size });
    await page.setContent(`<html><body style="margin:0;background:transparent">${svg.replace('<svg ', `<svg style="display:block;width:${size}px;height:${size}px" `)}</body></html>`);
    await page.screenshot({ path: `${outDir}/icon-${size}.png`, omitBackground: true });
  }
  await browser.close();
})();
