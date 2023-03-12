import fs from "fs";
import { CaughtLoomieModel, LoomieRarityModel } from "../models/mongoose.js";

function getRandomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

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

/**
 * Create a random team of 6 loomies to defend a gym
 * @param {*} commonLoomies Array of loomies with rarity "common"
 * @returns Array of the ids of the generated gym defenders
 */
export async function createRandomLoomieTeam(commonLoomies) {
  const commonLoomiesCopy = [...commonLoomies];

  let team = Array.from({ length: 6 }, () => {
    // Select a random loomie
    const randomIndex = Math.floor(Math.random() * commonLoomiesCopy.length);
    return commonLoomiesCopy[randomIndex];
  });

  // Replace the base stats names
  // (e.g. "base_attack" -> "attack")
  team = team.map((baseLoomie) => {
    return {
      // Shared attributes
      serial: baseLoomie.serial,
      name: baseLoomie.name,
      types: baseLoomie.types,
      rarity: baseLoomie.rarity,
      // Change names and reduce the stats to generate a weaker loomie
      hp: baseLoomie.base_hp - getRandomInt(15, 20),
      attack: baseLoomie.base_attack - getRandomInt(5, 10),
      defense: baseLoomie.base_defense - getRandomInt(0, 5),
      // The loomie has no owner, it just exists to protect the gym initially
      owner: null,
      // The loomie is busy defending the gym althought it's not owned by anyone
      is_busy: true,
    };
  });

  const inserted = await CaughtLoomieModel.insertMany(team);
  const insertedIds = inserted.map((loomie) => loomie._id);
  return insertedIds;
}
