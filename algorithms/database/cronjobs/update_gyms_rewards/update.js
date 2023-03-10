import dotenv from "dotenv";
import mongoose from "mongoose";
import {
  getRandomRewardsAmount,
  generateRewards,
  printRewards,
} from "./helpers.js";
import { GymModel, ItemModel, LoomBallModel } from "../../models/mongoose.js";

// Connect to MongoDB
dotenv.config();
mongoose.set("strictQuery", true);
mongoose.connect(process.env.MONGO_URI, { dbName: "loomies" });

// Debug variables
const PLAYERS_GENERATED_REWARDS = {};
const OWNERS_GENERATED_REWARDS = {};

// 1. Remove the current rewards and the users who claimed them
async function removeRewardsAndClaimers() {
  await GymModel.updateMany(
    {},
    // Just clear the arrays
    { current_rewards: [], rewards_claimed_by: [] }
  );
}

// 2. Generate & save the new rewards

async function generateNewRewards() {
  // Get all the items and loomballs
  const items = await ItemModel.find({});
  const loomballs = await LoomBallModel.find({});

  // Prepare the chances array
  const players_chances = [
    ...items.map((item) => ({
      result: { ...item._doc },
      chance: item.gym_reward_chance_player,
    })),
    ...loomballs.map((loomball) => ({
      result: { ...loomball._doc },
      chance: loomball.gym_reward_chance_player,
    })),
  ];

  const owners_chances = [
    ...items.map((item) => ({
      result: { ...item._doc },
      chance: item.gym_reward_chance_owner,
    })),
    ...loomballs.map((loomball) => ({
      result: { ...loomball._doc },
      chance: loomball.gym_reward_chance_owner,
    })),
  ];

  // Generate random rewards for each gym
  const gyms = await GymModel.find({});

  for await (const gym of gyms) {
    // Random rewards quantity [4-6]
    let min_rewards_qty = 4;
    let max_rewards_qty = 6;
    let rewards_qty = getRandomRewardsAmount(min_rewards_qty, max_rewards_qty);

    // Copy the chances array to avoid repeating the same reward
    let current_player_chances = [...players_chances];
    let current_owner_chances = [...owners_chances];
    const player_generated_rewards = [];
    const owner_generated_rewards = [];

    // Generate rewards for the players
    await generateRewards(
      3,
      5,
      players_chances,
      items,
      player_generated_rewards,
      PLAYERS_GENERATED_REWARDS
    );

    // Generate rewards for the owners
    await generateRewards(
      4,
      6,
      owners_chances,
      items,
      owner_generated_rewards,
      OWNERS_GENERATED_REWARDS
    );

    // Save the rewards
    await GymModel.updateOne(
      { _id: gym._id },
      {
        current_players_rewards: player_generated_rewards,
        current_owners_rewards: owner_generated_rewards,
      }
    );
  }
}

// 3. Run
async function run() {
  await removeRewardsAndClaimers();
  await generateNewRewards();
  printRewards(PLAYERS_GENERATED_REWARDS, "Rewards for players:");
  printRewards(OWNERS_GENERATED_REWARDS, "Rewards for owners:");
  mongoose.connection.close();
}

run();
