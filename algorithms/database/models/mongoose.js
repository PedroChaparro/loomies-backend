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
    loomies: [{ type: Schema.Types.ObjectId, ref: "wild_loomies" }],
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

const WildLoomieSchema = new Schema(
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
    hp: Number,
    attack: Number,
    defense: Number,
    zone_id: {
      type: Schema.Types.ObjectId,
      ref: "zones",
    },
    latitude: Number,
    longitude: Number,
    generated_at: Number,
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

const LoomBallsSchema = new Schema(
  {
    name: String,
    effective_until: Number,
    decay_until: Number,
    minimum_probability: {
      type: Number,
      min: 0,
      max: 1,
    },
  },
  { versionKey: false }
);

// Models
export const ZoneModel = model("zones", ZoneSchema);
export const GymModel = model("gyms", GymSchema);
export const LoomieTypeModel = model("loomie_types", LoomieTypeSchema);
export const LoomieRarityModel = model("loomie_rarities", LoomieRaritySchema);
export const BaseLoomieModel = model("base_loomies", BaseLoomieSchema);
export const WildLoomieModel = model("wild_loomies", WildLoomieSchema);
export const ItemModel = model("items", ItemsSchema);
export const LoomBallModel = model("loom_balls", LoomBallsSchema);
