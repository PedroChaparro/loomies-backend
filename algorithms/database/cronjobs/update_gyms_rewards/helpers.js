import Randomly from "weighted-randomly-select";

export function getRandomRewardsAmount(min, max) {
  return Math.floor(Math.random() * (max - min + 1) + min);
}

/**
 * Generate random rewards and save them in the given array and map
 * @param {Number} minQuantity Minimum number of rewards to generate
 * @param {Number} maxQuantity Maximum number of rewards to generate
 * @param {Number} quantity Number of rewards to generate
 * @param {*} chances Array of chances to generate rewards
 * @param {*} items Array of items to verify if the reward is an item or a loomball
 * @param {*} rewardsArr  Array of rewards to add the generated rewards
 * @param {*} rewardsMap  Map of rewards to add the generated rewards
 */
export async function generateRewards(
  minQuantity,
  maxQuantity,
  chances,
  items,
  rewardsArr,
  rewardsMap
) {
  // Random number of rewards
  const rewards_qty = getRandomRewardsAmount(minQuantity, maxQuantity);

  // Generate each reward
  for (let i = 0; i < rewards_qty; i++) {
    // Select random reward from the weighted array
    const selection = Randomly.select(chances);

    // Select the quantity based on the min and max reward quantity
    // properties of the selected item
    const min_qty = selection.min_reward_quantity;
    const max_qty = selection.max_reward_quantity;
    const qty = getRandomRewardsAmount(min_qty, max_qty);

    // Add to the rewards array
    const isItem = items.some((item) => item._doc._id === selection._id);
    rewardsArr.push({
      reward_collection: isItem ? "items" : "loom_balls",
      reward_id: selection._id,
      reward_quantity: qty,
    });

    // Add to the debug map
    rewardsMap[selection.name] = rewardsMap[selection.name]
      ? rewardsMap[selection.name] + qty
      : qty;

    // Remove from the chances array to avoid repeating the same reward
    chances = chances.filter((reward) => reward.result._id !== selection._id);
  }
}
