import dotenv from "dotenv";
import fs from "fs";
import mongoose from "mongoose";
import {
  ZoneModel,
  GymModel,
  BaseLoomieModel,
  ItemModel,
} from "./models/mongoose.js";

// Connect to MongoDB
dotenv.config();
mongoose.set("strictQuery", true);
mongoose.connect(process.env.MONGO_URI, { dbName: "loomies" });

// Read data from json files
const zones = JSON.parse(fs.readFileSync("../../data/zones.json"));
const gyms = JSON.parse(fs.readFileSync("../../data/places.json"));
const loomies = JSON.parse(fs.readFileSync("../../data/loomies.json"));
const items = JSON.parse(fs.readFileSync("../../data/items.json"));

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
  });

  await newZone.save();

  // Increment coordinates
  coordinates.x++;
}

console.log("Zones inserted: ", await ZoneModel.countDocuments());
console.log("Gyms inserted: ", await GymModel.countDocuments(), "\n");

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

  const newLoomie = new BaseLoomieModel({
    serial,
    name,
    types,
    rarity,
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

// Close connection
mongoose.connection.close();
