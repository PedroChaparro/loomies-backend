import dotenv from "dotenv";
import mongoose from "mongoose";
import {
  ZoneModel,
  GymModel,
  LoomieTypeModel,
  LoomieRarityModel,
  BaseLoomieModel,
  ItemModel,
  LoomBallModel,
} from "./models/mongoose.js";
import { readJsonFromDataFolder } from "./utils/utils.js";

// Connect to MongoDB
dotenv.config();
mongoose.set("strictQuery", true);
mongoose.connect(process.env.MONGO_URI, { dbName: "loomies" });

// Read data from json files
const zones = readJsonFromDataFolder("zones");
const gyms = readJsonFromDataFolder("places");
const loomies = readJsonFromDataFolder("loomies");
const items = readJsonFromDataFolder("items");
const loomieTypes = readJsonFromDataFolder("loomies_types");
const loomieRarities = readJsonFromDataFolder("loomies_rarities");
const loomballs = readJsonFromDataFolder("loomballs");

// --- Zones and Gyms ---
console.log("🏟️ Inserting gyms and zones...");
const coordinates = { x: 0, y: 0 };
let currentLongitude;

console.log("Expected zones: ", zones.length);
console.log("Expected gyms: ", gyms.length);

for await (const zone of zones) {
  let GymMongoId;

  // Initialize currentLongitude
  if (!currentLongitude) currentLongitude = zone.bottomFrontier;

  // Increment coordinates when longitude changes (New row)
  if (currentLongitude !== zone.bottomFrontier) {
    currentLongitude = zone.bottomFrontier;
    coordinates.x = 0;
    coordinates.y++;
  }

  // Get the zone's gym
  const gym = gyms.findIndex((gym) => gym.zoneIdentifier === zone.identifier);

  // Insert the gym into mongodb and get the id
  if (gym !== -1) {
    const { name, latitude, longitude } = gyms[gym];
    const newGym = new GymModel({ name, latitude, longitude });
    const { _id } = await newGym.save();
    GymMongoId = _id;
  }

  // Insert zone with the gym id
  const { leftFrontier, rightFrontier, topFrontier, bottomFrontier, number } =
    zone;

  const newZone = new ZoneModel({
    leftFrontier,
    rightFrontier,
    topFrontier,
    bottomFrontier,
    number,
    coordinates: `${coordinates.x},${coordinates.y}`,
    gym: GymMongoId ? GymMongoId : null,
    loomies: [], // Empty loomies array
  });

  await newZone.save();

  // Increment coordinates
  coordinates.x++;
}

console.log("Zones inserted: ", await ZoneModel.countDocuments());
console.log("Gyms inserted: ", await GymModel.countDocuments(), "\n");

// --- Loomies types ---
console.log("✨ Inserting loomie types...");
const globalLoomiesTypesIds = [];

// 1. Insert loomie types without the strong_against attribute
for await (const loomieType of loomieTypes) {
  const { name } = loomieType;
  const newLoomieType = new LoomieTypeModel({ name });

  // Save the id to populate the strong_against attribute later
  const { _id } = await newLoomieType.save();
  globalLoomiesTypesIds.push({
    name: loomieType.name,
    id: _id,
  });
}

// 2. Update strong_against attribute
for await (const loomieType of loomieTypes) {
  const { name, strong_against } = loomieType;
  const strongAgainstIds = [];

  // Get the current loomie
  const currentLoomie = globalLoomiesTypesIds.find(
    (loomie_type) => loomie_type.name === name
  );

  if (!currentLoomie) {
    console.log("⚠️ Loomie was not found:", name);
    continue;
  }

  // Get the ids of the strong_against loomies
  for await (const strongAgainst of strong_against) {
    const strongAgainstId = globalLoomiesTypesIds.find(
      (loomie_type) => loomie_type.name === strongAgainst
    );

    if (!strongAgainstId) {
      console.log(
        `⚠️ Strong against loomie was not found: ${currentLoomie.name} --> ${strongAgainst}`
      );
      continue;
    }

    strongAgainstIds.push(strongAgainstId.id);
  }

  // Update the current loomie
  await LoomieTypeModel.updateOne(
    { _id: currentLoomie.id },
    { strong_against: strongAgainstIds }
  );
}

console.log(
  "Inserted loomie types: ",
  await LoomieTypeModel.countDocuments(),
  "\n"
);

// --- Loomies rarities ---
console.log("📊 Inserting loomie rarities...");
const globalLoomiesRaritiesIds = [];

for await (const loomieRarity of loomieRarities) {
  const { name, spawn_chance } = loomieRarity;
  const newLoomieRarity = new LoomieRarityModel({ name, spawn_chance });

  // Save the id to populate the loomies.rarity attribute later
  const { _id } = await newLoomieRarity.save();
  globalLoomiesRaritiesIds.push({
    name: loomieRarity.name,
    id: _id,
  });
}

console.log(
  "Inserted loomie rarities: ",
  await LoomieRarityModel.countDocuments(),
  "\n"
);

// --- Loomies ---
console.log("🐄 Inserting loomies...");

for await (const loomie of loomies) {
  const BASE_ATTRIBUTES = {
    hp: 100,
    deffense: 10,
    attack: 20,
  };

  const { serial, name, types, rarity, extra_hp, extra_def, extra_atk } =
    loomie;

  // Get the ids of the types
  const typesIds = [];

  for await (const type of types) {
    const typeId = globalLoomiesTypesIds.find(
      (loomie_type) => loomie_type.name === type
    );

    if (!typeId) {
      console.log("⚠️ Loomie type was not found:", type);
      continue;
    }

    typesIds.push(typeId.id);
  }

  // Get the id of the rarity
  const rarityId = globalLoomiesRaritiesIds.find(
    (loomie_rarity) => loomie_rarity.name === rarity
  );

  if (!rarityId) {
    console.log("⚠️ Loomie rarity was not found:", rarity);
    continue;
  }

  const newLoomie = new BaseLoomieModel({
    serial,
    name,
    types: typesIds,
    rarity: rarityId.id,
    base_hp: BASE_ATTRIBUTES.hp + extra_hp,
    base_attack: BASE_ATTRIBUTES.attack + extra_atk,
    base_defense: BASE_ATTRIBUTES.deffense + extra_def,
  });

  await newLoomie.save();
}

console.log("Inserted loomies: ", await BaseLoomieModel.countDocuments(), "\n");

// --- Items ---
console.log("📦 Inserting items...");

for await (const item of items) {
  const { name, description, target, is_combat_item } = item;

  const newItem = new ItemModel({
    name,
    description,
    target,
    is_combat_item,
  });

  await newItem.save();
}

console.log("Inserted items: ", await ItemModel.countDocuments(), "\n");

// --- Loomballs ---
console.log("🎱 Inserting loomballs...");

for await (const loomball of loomballs) {
  const { name, effective_until, decay_until, minimum_probability } = loomball;

  const newLoomball = new LoomBallModel({
    name,
    effective_until,
    decay_until,
    minimum_probability,
  });

  await newLoomball.save();
}

console.log("Inserted loomballs: ", await LoomBallModel.countDocuments(), "\n");

// Close connection
await ZoneModel.ensureIndexes();
mongoose.connection.close();
