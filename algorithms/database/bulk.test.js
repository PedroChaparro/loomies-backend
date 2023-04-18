import dotenv from "dotenv";
import mongoose from "mongoose";
import { it, describe, expect } from "vitest";
import {
  BaseLoomieModel,
  GymModel,
  ItemModel,
  LoomBallModel,
  LoomieRarityModel,
  LoomieTypeModel,
  ZoneModel,
} from "./models/mongoose";
import { readJsonFromDataFolder } from "./utils/utils";

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

describe("Test loomies rarities", () => {
  it("Should have the corrects spawn chances", async () => {
    const loomieRaritiesDocuments = await LoomieRarityModel.find();

    for await (const loomieRarityDoc of loomieRaritiesDocuments) {
      // Get the loomie spawn chance from the json file
      const { spawn_chance } = loomieRarities.find(
        (loomie_rarity) => loomie_rarity.name === loomieRarityDoc.name
      );

      // Validate spawn chance
      expect(spawn_chance).toBe(loomieRarityDoc.spawn_chance);
    }
  });
});

describe("Test Loomballs", () => {
  it("Shold have all the expected loomballs", async () => {
    const loomballsDocuments = await LoomBallModel.find();

    // Validate count
    expect(loomballsDocuments.length).toBe(loomballs.length);

    for await (const loomballDoc of loomballsDocuments) {
      // Get the loomball from the json file
      const loomball = loomballs.find(
        (loomball) => loomball.name === loomballDoc.name
      );

      // Validate loomball
      expect(loomball).not.toBe(undefined);
      expect(loomballDoc.name).toBe(loomball.name);
      expect(loomballDoc.effective_until).toBe(loomball.effective_until);
      expect(loomballDoc.decay_until).toBe(loomball.decay_until);
      expect(loomballDoc.minimum_probability).toBe(
        loomball.minimum_probability
      );
    }
  });
});
