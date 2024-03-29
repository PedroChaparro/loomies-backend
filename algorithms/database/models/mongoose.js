import { Schema, model } from "mongoose";

// -- --- --- --- ---
// Schemas
const ZoneSchema = new Schema(
  {
    leftFrontier: Number,
    rightFrontier: Number,
    topFrontier: Number,
    bottomFrontier: Number,
    number: Number,
    coordinates: String,
    gyms: [{ type: Schema.Types.ObjectId, ref: "gyms" }],
    loomies: [{ type: Schema.Types.ObjectId, ref: "wild_loomies" }],
  },
  { versionKey: false }
);

// Create hash and unique index for coordinates
ZoneSchema.set("autoIndex", false);
ZoneSchema.index({ coordinates: "hashed" });

// Create a schema for the rewards that can be claimed by players and gym owners
const sharedRewardSchema = {
  type: [
    {
      _id: false,
      reward_collection: {
        type: String,
        // Gym rewards can be items or loomballs
        enum: ["items", "loom_balls"],
      },
      reward_id: {
        type: Schema.Types.ObjectId,
        // Dynamically reference the correct collection
        refPath: "current_rewards.reward_type",
      },
      reward_quantity: Number,
    },
  ],
};

const GymSchema = new Schema(
  {
    latitude: Number,
    longitude: Number,
    name: String,
    owner: { type: Schema.Types.ObjectId, ref: "users" },
    protectors: [{ type: Schema.Types.ObjectId, ref: "caught_loomies" }],
    current_players_rewards: sharedRewardSchema,
    current_owners_rewards: sharedRewardSchema,
    rewards_claimed_by: [{ type: Schema.Types.ObjectId, ref: "users" }],
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
    serial: {
      type: Number,
      unique: true,
    },
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

const sharedLoomieAttributes = {
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
};

const WildLoomieSchema = new Schema(
  {
    ...sharedLoomieAttributes,
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

const CaughtLoomieSchema = new Schema(
  {
    // The caught loomie is a copy of the wild loomie
    ...sharedLoomieAttributes,
    level: {
      type: Number,
      min: 1,
      default: 1,
    },
    experience: {
      type: Number,
      min: 0,
      default: 0,
    },
    // The caught loomie can be busy if it's
    is_busy: Boolean,
    // But also has a reference to the user that caught it
    owner: { type: Schema.Types.ObjectId, ref: "users" },
  },
  { versionKey: false }
);

const ItemsSchema = new Schema(
  {
    name: String,
    serial: {
      type: Number,
      unique: true,
    },
    description: String,
    target: {
      type: String,
      enum: ["Loomie"], // Currently items only target loomies
    },
    is_combat_item: Boolean,
    // Probability to appear as a gym reward
    gym_reward_chance_player: {
      type: Number,
      min: 0,
      max: 1,
    },
    gym_reward_chance_owner: {
      type: Number,
      min: 0,
      max: 1,
    },
    min_reward_quantity: Number,
    max_reward_quantity: Number,
  },
  { versionKey: false }
);

const LoomBallsSchema = new Schema(
  {
    name: String,
    serial: {
      type: Number,
      unique: true,
    },
    effective_until: Number,
    decay_until: Number,
    minimum_probability: {
      type: Number,
      min: 0,
      max: 1,
    },
    // Probability to appear as a gym reward
    gym_reward_chance_player: {
      type: Number,
      min: 0,
      max: 1,
    },
    gym_reward_chance_owner: {
      type: Number,
      min: 0,
      max: 1,
    },
    min_reward_quantity: Number,
    max_reward_quantity: Number,
  },
  { versionKey: false }
);

// user items
const userItemSchema = {
  type: [
    {
      _id: false,
      item_collection: {
        type: String,
        enum: ["items", "loom_balls"],
      },
      item_id: {
        type: Schema.Types.ObjectId,
        // Dynamically reference the correct collection
        refPath: "items._id",
      },
      item_quantity: Number,
    },
  ],
};

const UserSchema = new Schema(
  {
    username: String,
    email: String,
    password: String,
    items: userItemSchema,
    loomies: [{ type: Schema.Types.ObjectId, ref: "caught_loomies" }],
    loomie_team: [{ type: Schema.Types.ObjectId, ref: "caught_loomies" }],
    isVerified: Boolean,
    currentLoomiesGenerationTimeout: Number,
    lastLoomieGenerationTime: Number,
  },
  { versionKey: false }
);

// -- --- --- --- ---
// Models

// Zone & Gyms
export const ZoneModel = model("zones", ZoneSchema);
export const GymModel = model("gyms", GymSchema);
// Loomies
export const LoomieTypeModel = model("loomie_types", LoomieTypeSchema);
export const LoomieRarityModel = model("loomie_rarities", LoomieRaritySchema);
export const BaseLoomieModel = model("base_loomies", BaseLoomieSchema);
export const WildLoomieModel = model("wild_loomies", WildLoomieSchema);
export const CaughtLoomieModel = model("caught_loomies", CaughtLoomieSchema);
// Collectionables
export const ItemModel = model("items", ItemsSchema);
export const LoomBallModel = model("loom_balls", LoomBallsSchema);
// User
export const UserModel = model("users", UserSchema);
