import fs from "fs";

export function readJsonFromDataFolder(name) {
  const path = `../../data/${name}.json`;
  try {
    const data = fs.readFileSync(path);
    return JSON.parse(data);
  } catch (err) {
    console.log(`Error reading ${path}:`);
    console.log(err);
    process.exit(1);
  }
}
