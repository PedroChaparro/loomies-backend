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
describe.concurrent("Testing documents count", () => {
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

describe(
  "Testing zones and gyms",
  () => {
    it("Should have the same number of zones and gyms", async () => {
      const zonesCount = await ZoneModel.countDocuments();
      const gymsCount = await GymModel.countDocuments();
      expect(zonesCount).toBe(gymsCount);
    });

    it("Should not have any zone without gym", async () => {
      const zonesWithoutGym = await ZoneModel.find({ gym: null });
      expect(zonesWithoutGym.length).toBe(0);
    });

    it("Should not have zones with the zame gym", async () => {
      const Gyms = await GymModel.find();

      for await (const gym of Gyms) {
        const zones = await ZoneModel.find({ gym: gym._id });
        expect(zones.length).toBe(1);
      }
    });
  },
  // 15 seconds timeout
  { timeout: 15000 }
);

describe("Testing loomies types", () => {
  it("Should have the corrects strong against ids", async () => {
    const loomieTypesDocuments = await LoomieTypeModel.find().populate(
      "strong_against"
    );

    for await (const loomieTypeDoc of loomieTypesDocuments) {
      // Get the loomie strong against names from the json file
      const { strong_against } = loomieTypes.find(
        (loomie_type) => loomie_type.name === loomieTypeDoc.name
      ); // ["Fire", "Water"...]

      // Validate count
      expect(strong_against.length).toBe(loomieTypeDoc.strong_against.length);

      // Every name in the json file should be in the document
      const isEveryNameIncluded = strong_against.every((name) => {
        return loomieTypeDoc.strong_against.some((type) => type.name === name);
      });

      expect(isEveryNameIncluded).toBe(true);
    }
  });
});
