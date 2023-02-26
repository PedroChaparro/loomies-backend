import { Schema, model } from "mongoose";

// Schemas
const ZoneSchema = new Schema(
  {
    leftFrontier: Number,
    rightFrontier: Number,
    topFrontier: Number,
    bottomFrontier: Number,
    number: Number,
    coordinates: String,
    gym: { type: Schema.Types.ObjectId, ref: "Gym" },
  },
  { versionKey: false }
);

// Create hash and unique index for coordinates
ZoneSchema.set("autoIndex", false);
ZoneSchema.index({ coordinates: "hashed" });

const GymSchema = new Schema(
  {
    latitude: Number,
    longitude: Number,
    name: String,
  },
  { versionKey: false }
);

const LoomieTypeSchema = new Schema(
  {
    name: String,
    strong_against: {
      // Types that this rarity is strong against
      // Reference another document in the same collection
      type: [Schema.Types.ObjectId],
      ref: "loomie_types",
    },
  },
  { versionKey: false }
);

const LoomieRaritySchema = new Schema(
  {
    name: String,
    spawn_chance: Number,
  },
  { versionKey: false }
);

const BaseLoomieSchema = new Schema(
  {
    serial: Number,
    name: String,
    types: {
      type: [Schema.Types.ObjectId],
      ref: "loomie_types",
    },
    rarity: {
      type: Schema.Types.ObjectId,
      ref: "loomie_rarities",
    },
    base_hp: Number,
    base_attack: Number,
    base_defense: Number,
  },
  { versionKey: false }
);

const ItemsSchema = new Schema(
  {
    name: String,
    description: String,
    target: {
      type: String,
      enum: ["Loomie"], // Currently items only target loomies
    },
    is_combat_item: Boolean,
  },
  { versionKey: false }
);

// Models
export const ZoneModel = model("zones", ZoneSchema);
export const GymModel = model("gyms", GymSchema);
export const LoomieTypeModel = model("loomie_types", LoomieTypeSchema);
export const LoomieRarityModel = model("loomie_rarities", LoomieRaritySchema);
export const BaseLoomieModel = model("base_loomies", BaseLoomieSchema);
export const ItemModel = model("items", ItemsSchema);