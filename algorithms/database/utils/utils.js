import fs from "fs";
import {
  CaughtLoomieModel,
  LoomieRarityModel,
  UserModel,
} from "../models/mongoose.js";

export function getRandomInt(min, max) {
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
      level: getRandomInt(14, 24),
      name: baseLoomie.name,
      types: baseLoomie.types,
      rarity: baseLoomie.rarity,
      // Change names and reduce the stats to generate a weaker loomie
      hp: baseLoomie.base_hp + getRandomInt(-2, 4),
      attack: baseLoomie.base_attack + getRandomInt(-2, 4),
      defense: baseLoomie.base_defense + getRandomInt(-2, 4),
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

export async function createHardcoreLoomieTeam(rareLoomies, normalLoomies) {
  const possibleLoomies = [...rareLoomies, ...normalLoomies];

  // Take the first 6 loomies
  let team = possibleLoomies.slice(0, 6);

  // Replace the base stats names
  team = team.map((baseLoomie) => {
    return {
      serial: baseLoomie.serial,
      name: baseLoomie.name,
      types: baseLoomie.types,
      rarity: baseLoomie.rarity,
      level: getRandomInt(40, 50),
      hp: baseLoomie.base_hp + getRandomInt(0, 5),
      attack: baseLoomie.base_attack + getRandomInt(0, 5),
      defense: baseLoomie.base_defense + getRandomInt(0, 5),
      owner: null,
      is_busy: true,
    };
  });

  const inserted = await CaughtLoomieModel.insertMany(team);
  const insertedIds = inserted.map((loomie) => loomie._id);
  return insertedIds;
}

/**
 * Instantiates all existing Loomies and adds them to a player
 * @param {*} possibleLoomies Array of Loomies
 * @param {string} owner Mongo object id of owner player
 */
export async function giveAllLoomies(possibleLoomies, ownerId) {
  try {
    // Replace the base stats names
    const newLoomies = possibleLoomies.map((baseLoomie) => {
      return {
        serial: baseLoomie.serial,
        name: baseLoomie.name,
        types: baseLoomie.types,
        rarity: baseLoomie.rarity,
        level: getRandomInt(1, 30),
        hp: baseLoomie.base_hp + getRandomInt(0, 5),
        attack: baseLoomie.base_attack + getRandomInt(0, 5),
        defense: baseLoomie.base_defense + getRandomInt(0, 5),
        owner: ownerId,
        is_busy: false,
      };
    });

    const inserted = await CaughtLoomieModel.insertMany(newLoomies);
    const insertedIds = inserted.map((loomie) => loomie._id);

    // add to ids to owner

    await UserModel.findOneAndUpdate(
      { _id: ownerId },
      {
        $push: {
          loomies: insertedIds,
        },
      }
    );
  } catch (e) {
    console.log("Here goes the error");
    console.error(e);
  }
}

/**
 *
 * @param {number} latitude The latitude of the gym
 * @param {number} longitude The longitude of the gym
 * @returns The local coordinates of the zone where the gym is located
 */
export function getZoneCoordinatesFromGPS(latitude, longitude) {
  const initialLatitude = 6.9595;
  const initialLongitude = -73.1696;
  const zoneSize = 0.0035;

  const x = Math.floor((longitude - initialLongitude) / zoneSize);
  const y = Math.floor((latitude - initialLatitude) / zoneSize);
  return { x, y };
}
