import dotenv from "dotenv";
import mongoose from "mongoose";
import Randomly from "weighted-randomly-select";
import { GymModel, ItemModel, LoomBallModel } from "../../models/mongoose.js";

// Connect to MongoDB
dotenv.config();
mongoose.set("strictQuery", true);
mongoose.connect(process.env.MONGO_URI, { dbName: "loomies" });

// Debug variables
const CHOOSED_REWARDS = {};

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
  const chances = [
    ...items.map((item) => ({
      result: { ...item._doc },
      chance: item.gym_reward_chance_player,
    })),
    ...loomballs.map((loomball) => ({
      result: { ...loomball._doc },
      chance: loomball.gym_reward_chance_player,
    })),
  ];

  // Generate random rewards for each gym
  const gyms = await GymModel.find({});

  for await (const gym of gyms) {
    // Random rewards quantity [4-6]
    const min_rewards_qty = 4;
    const max_rewards_qty = 6;
    const rewards_qty = Math.floor(
      Math.random() * (max_rewards_qty - min_rewards_qty + 1) + min_rewards_qty
    );

    // Copy the chances array to avoid repeating the same reward
    let current_chances = [...chances];
    const current_generated_rewards = [];

    for (let i = 0; i < rewards_qty; i++) {
      // Select
      const selection = Randomly.select(current_chances);
      const min_qty = selection.min_reward_quantity;
      const max_qty = selection.max_reward_quantity;
      const qty = Math.floor(Math.random() * (max_qty - min_qty + 1) + min_qty);

      // Add to the gym rewards
      const isItem = items.some((item) => item._doc._id === selection._id);

      current_generated_rewards.push({
        reward_collection: isItem ? "items" : "loom_balls",
        reward_id: selection._id,
        reward_quantity: qty,
      });

      // Add to the debug variable
      CHOOSED_REWARDS[selection.name] = CHOOSED_REWARDS[selection.name]
        ? CHOOSED_REWARDS[selection.name] + qty
        : qty;

      // Remove from the chances array
      current_chances = current_chances.filter(
        (reward) => reward.result._id !== selection._id
      );
    }

    // Save the rewards
    await GymModel.updateOne(
      { _id: gym._id },
      { current_rewards: current_generated_rewards }
    );
  }
}

// 3. Run
async function run() {
  await removeRewardsAndClaimers();
  await generateNewRewards();
  console.log("The generated rewards are: (name: quantity): ");

  const GENERATED_ARR = Object.entries(CHOOSED_REWARDS);
  GENERATED_ARR.sort((a, b) => b[1] - a[1]);
  console.table(GENERATED_ARR);

  mongoose.connection.close();
}

run();
