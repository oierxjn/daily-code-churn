const { execFileSync } = require("child_process");
const fs = require("fs");
const path = require("path");

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
const commitBranchInput = getInput("commit_branch");

function getEnv(name) {
  const v = process.env[name];
  return (v === undefined || v === "") ? undefined : v;
}

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
const bin = path.join(__dirname, "dist", "daily-code-churn-linux-amd64");
if (!fs.existsSync(bin)) {
  console.error(`Binary not found: ${bin}. Did you forget to build and commit dist/?`);
  process.exit(1);
}
if (process.platform !== "win32") {
  try {
    fs.chmodSync(bin, 0o755);
  } catch (err) {
    console.error(`Failed to chmod ${bin}:`, err.message || err);
    process.exit(1);
  }
}

const args = [];
if (days) args.push("-days", days);
if (out) args.push("-out", out);
if (width) args.push("-width", width);
if (height) args.push("-height", height);
if (branch) args.push("-branch", branch);

console.log("Running:", bin, args.join(" "));
execFileSync(bin, args, { stdio: "inherit" });

const outPath = out || getEnv("CHURN_OUT") || "daily-churn.svg";

/**
 * 检查指定路径的文件是否被Git修改
 * @param {string} path - 需要检查的文件或目录路径
 * @returns {boolean} - 如果文件被修改则返回true，否则返回false；出错时也返回false
 */
function gitChanged(path) {
  try {
    // 使用git status命令检查文件状态
    // --porcelain参数以简洁格式输出状态信息
    // --确保只检查指定路径
    const out = execFileSync("git", ["status", "--porcelain", "--", path], { stdio: "pipe" });
    // 将输出转为字符串并去除首尾空白，如果结果不为空则表示文件有变更
    return out.toString().trim() !== "";
  } catch (err) {
    console.error("git status failed:", err.message || err);
    return false;
  }
}

if (!fs.existsSync(outPath)) {
  console.error(`Output not found: ${outPath}`);
  process.exit(1);
}

if (!gitChanged(outPath)) {
  console.log(`No changes in ${outPath}`);
  process.exit(0);
}

try {
  execFileSync("git", ["add", outPath], { stdio: "inherit" });
  execFileSync("git", ["config", "user.name", process.env.GITHUB_ACTOR || "github-actions[bot]"], { stdio: "inherit" });
  execFileSync("git", ["config", "user.email", `${process.env.GITHUB_ACTOR || "github-actions[bot]"}@users.noreply.github.com`], { stdio: "inherit" });
  execFileSync("git", ["commit", "-m", "chore: update churn output"], { stdio: "inherit" });
  let branch = commitBranchInput || getEnv("CHURN_COMMIT_BRANCH") || process.env.GITHUB_REF_NAME;
  if (!branch) {
    try {
      const out = execFileSync("git", ["symbolic-ref", "--short", "HEAD"], { stdio: "pipe" });
      branch = out.toString().trim();
    } catch {}
  }
  if (!branch) {
    console.error("No target branch specified. Set CHURN_COMMIT_BRANCH in your workflow.");
    process.exit(1);
  }
  execFileSync("git", ["push", "origin", `HEAD:${branch}`], { stdio: "inherit" });
} catch (err) {
  console.error("git commit/push failed:", err.message || err);
  process.exit(1);
}
