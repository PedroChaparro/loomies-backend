import dotenv from "dotenv";
import mongoose from "mongoose";
import Randomly from "weighted-randomly-select";
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
function getRandomRewardsAmount(min, max) {
  return Math.floor(Math.random() * (max - min + 1) + min);
}

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

    // -- Generate rewards for the players
    for (let i = 0; i < rewards_qty; i++) {
      const player_sel = Randomly.select(current_player_chances);
      const min_qty = player_sel.min_reward_quantity;
      const max_qty = player_sel.max_reward_quantity;
      const qty = Math.floor(Math.random() * (max_qty - min_qty + 1) + min_qty);

      // Add to the gym rewards
      const isItem = items.some((item) => item._doc._id === player_sel._id);

      player_generated_rewards.push({
        reward_collection: isItem ? "items" : "loom_balls",
        reward_id: player_sel._id,
        reward_quantity: qty,
      });

      // Add to the debug variable
      PLAYERS_GENERATED_REWARDS[player_sel.name] = PLAYERS_GENERATED_REWARDS[
        player_sel.name
      ]
        ? PLAYERS_GENERATED_REWARDS[player_sel.name] + qty
        : qty;

      // Remove from the chances array
      current_player_chances = current_player_chances.filter(
        (reward) => reward.result._id !== player_sel._id
      );
    }

    // -- Generate rewards for the owners
    min_rewards_qty = 5;
    max_rewards_qty = 7;
    rewards_qty = getRandomRewardsAmount(min_rewards_qty, max_rewards_qty);

    for (let i = 0; i < rewards_qty; i++) {
      const owner_sel = Randomly.select(current_owner_chances);
      const min_qty = owner_sel.min_reward_quantity;
      const max_qty = owner_sel.max_reward_quantity;
      const qty = Math.floor(Math.random() * (max_qty - min_qty + 1) + min_qty);

      // Add to the gym rewards
      const isItem = items.some((item) => item._doc._id === owner_sel._id);

      owner_generated_rewards.push({
        reward_collection: isItem ? "items" : "loom_balls",
        reward_id: owner_sel._id,
        reward_quantity: qty,
      });

      // Add to the debug variable
      OWNERS_GENERATED_REWARDS[owner_sel.name] = OWNERS_GENERATED_REWARDS[
        owner_sel.name
      ]
        ? OWNERS_GENERATED_REWARDS[owner_sel.name] + qty
        : qty;

      // Remove from the chances array
      current_owner_chances = current_owner_chances.filter(
        (reward) => reward.result._id !== owner_sel._id
      );
    }

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
  console.log("The generated rewards are: (name: quantity): ");

  const USER_GENERATED_ARR = Object.entries(PLAYERS_GENERATED_REWARDS);
  USER_GENERATED_ARR.sort((a, b) => b[1] - a[1]);
  console.log("Generated rewards for the players:");
  console.table(USER_GENERATED_ARR);

  const OWNER_GENERATED_ARR = Object.entries(OWNERS_GENERATED_REWARDS);
  OWNER_GENERATED_ARR.sort((a, b) => b[1] - a[1]);
  console.log("Generated rewards for the owners:");
  console.table(OWNER_GENERATED_ARR);

  mongoose.connection.close();
}

run();
