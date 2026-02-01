const { execFileSync } = require("child_process");
const fs = require("fs");

/**
 * 获取输入环境变量的值
 * @param {string} name - 输入参数的名称
 * @returns {string|undefined} 返回环境变量的值，如果不存在或为空则返回undefined
 */
function getInput(name) {
  const v = process.env[`INPUT_${name.toUpperCase()}`];
  return (v === undefined || v === "") ? undefined : v;
}

const days = getInput("days");
const branch = getInput("branch");
const out = getInput("out");
const width = getInput("width");
const height = getInput("height");

function platformName() {
  switch (process.platform) {
    case "linux":
      return "linux";
    case "darwin":
      return "darwin";
    case "win32":
      return "windows";
    default:
      return undefined;
  }
}

function archName() {
  switch (process.arch) {
    case "x64":
      return "amd64";
    case "arm64":
      return "arm64";
    default:
      return undefined;
  }
}

const plat = platformName();
const arch = archName();
if (!plat || !arch) {
  console.error(`Unsupported platform/arch: ${process.platform}/${process.arch}`);
  process.exit(1);
}

const ext = process.platform === "win32" ? ".exe" : "";
const bin = `./dist/daily-code-churn-${plat}-${arch}${ext}`;
if (!fs.existsSync(bin)) {
  console.error(`Binary not found: ${bin}. Did you forget to build and commit dist/?`);
  process.exit(1);
}

const args = [];
if (days) args.push("-days", days);
if (out) args.push("-out", out);
if (width) args.push("-width", width);
if (height) args.push("-height", height);
if (branch) args.push("-branch", branch);

console.log("Running:", bin, args.join(" "));
execFileSync(bin, args, { stdio: "inherit" });
