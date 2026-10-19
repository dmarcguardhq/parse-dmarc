// Loads each URL in headless Chrome and fails unless the dashboard rendered
// and every request to the proxy stayed under the URL's path prefix with a
// non-error status. Requests to other origins (web fonts) are ignored.
import puppeteer from "puppeteer";

var browser = await puppeteer.launch({ args: ["--no-sandbox"] });
var failed = false;

for (var arg of process.argv.slice(2)) {
  var start = new URL(arg);
  // "/dmarc" for both http://nginx/dmarc and http://nginx/dmarc/index.html.
  var prefix = "/" + start.pathname.split("/")[1];
  var page = await browser.newPage();
  var seen = [];
  page.on("response", function onResponse(r) {
    seen.push({ url: new URL(r.url()), status: r.status() });
  });
  page.on("requestfailed", function onFailed(r) {
    seen.push({ url: new URL(r.url()), status: "failed" });
  });

  await page.goto(start.href, { waitUntil: "networkidle0" });
  var text = await page.$eval("#app", function appText(el) {
    return el.innerText;
  });
  await page.close();

  var problems = [];
  var own = seen.filter(function sameOrigin(r) {
    return r.url.origin === start.origin;
  });
  console.log(`${start.href}`);
  for (var r of own) {
    var path = r.url.pathname + r.url.search;
    console.log(`  ${r.status} ${path}`);
    var underPrefix =
      r.url.pathname === prefix || r.url.pathname.startsWith(prefix + "/");
    if (!underPrefix) problems.push(`outside ${prefix}/: ${path}`);
    if (r.status === "failed" || r.status >= 400) {
      problems.push(`${r.status}: ${path}`);
    }
  }
  for (var want of ["/assets/", "/api/statistics", "/api/reports"]) {
    var hit = own.some(function requested(r) {
      return r.url.pathname.startsWith(prefix + want);
    });
    if (!hit) problems.push(`never requested ${prefix}${want}`);
  }
  // Rendered by the Vue bundle, absent from the static index.html.
  if (!text.includes("DNS Generator")) {
    problems.push("dashboard did not render");
  }

  // A blocked script shows up both as a 404 response and a failed request.
  for (var p of new Set(problems)) console.log(`  FAIL ${p}`);
  failed = failed || problems.length > 0;
}

await browser.close();
console.log(failed ? "FAIL" : "PASS");
process.exit(failed ? 1 : 0);
