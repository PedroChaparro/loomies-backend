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

const BaseLoomieSchema = new Schema(
  {
    serial: Number,
    name: String,
    types: {
      type: [String],
      enum: [
        "Water",
        "Fire",
        "Plant",
        "Flying",
        "Psychic",
        "Bug",
        "Poison",
        "Electric",
        "Rock",
        "Iron",
      ],
    },
    rarity: {
      type: String,
      enum: ["Common", "Normal", "Rare"],
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
export const BaseLoomieModel = model("base_loomies", BaseLoomieSchema);
export const ItemModel = model("items", ItemsSchema);
