import dotenv from "dotenv";
import fs from "fs";
import mongoose from "mongoose";
import { it, describe, expect } from "vitest";
import {
  BaseLoomieModel,
  GymModel,
  ItemModel,
  LoomieRarityModel,
  LoomieTypeModel,
  ZoneModel,
} from "./models/mongoose";

// Connect to MongoDB
dotenv.config();
mongoose.set("strictQuery", true);
mongoose.connect(process.env.MONGO_URI, { dbName: "loomies" });

// Read data from json files
const zones = JSON.parse(fs.readFileSync("../../data/zones.json"));
const gyms = JSON.parse(fs.readFileSync("../../data/places.json"));
const loomies = JSON.parse(fs.readFileSync("../../data/loomies.json"));
const items = JSON.parse(fs.readFileSync("../../data/items.json"));
const loomieTypes = JSON.parse(
  fs.readFileSync("../../data/loomies_types.json")
);
const loomieRarities = JSON.parse(
  fs.readFileSync("../../data/loomies_rarities.json")
);

// --- Tests ---
describe("Testing documents count", () => {
  it(`Should have ${zones.length} zones`, async () => {
    expect(zones.length).toBe(await ZoneModel.countDocuments());
  });

  it(`Should have ${gyms.length} gyms`, async () => {
    expect(gyms.length).toBe(await GymModel.countDocuments());
  });

  it(`Should have ${loomieTypes.length} loomie types`, async () => {
    expect(loomieTypes.length).toBe(await LoomieTypeModel.countDocuments());
  });

  it(`Should have ${loomieRarities.length} loomie rarities`, async () => {
    expect(loomieRarities.length).toBe(
      await LoomieRarityModel.countDocuments()
    );
  });

  it(`Should have ${loomies.length} loomies`, async () => {
    expect(loomies.length).toBe(await BaseLoomieModel.countDocuments());
  });

  it(`Should have ${items.length} items`, async () => {
    expect(items.length).toBe(await ItemModel.countDocuments());
  });
});
