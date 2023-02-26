import dotenv from "dotenv";
import mongoose from "mongoose";
import { WildLoomieModel, ZoneModel } from "./models/mongoose.js";

// Connect to MongoDB
dotenv.config();
mongoose.set("strictQuery", true);
mongoose.connect(process.env.MONGO_URI, { dbName: "loomies" });

// Get command arguments
let all, outdated;
all = process.argv.includes("--all");
outdated = process.argv.includes("--outdated");

// Count documents to compare after the script
const wildLoomiesBefore = await WildLoomieModel.find({}).countDocuments();

// Remove loomies from the wild_loomies collection and clean the loomies array in the zones collection
if (all) {
  console.log("🔥 Removing all loomies from database...");
  await ZoneModel.updateMany({}, { loomies: [] });
  await WildLoomieModel.deleteMany({});
  console.log("🔥 Done!");
}

// Remove outdated loomies from the wild_loomies collection and remove them from the loomies array in the zones collection
if (outdated) {
  console.log("🔥 Removing outdated loomies from database...");

  const currentUnixTime = Math.trunc(new Date() / 1000);
  const outdatedTimeout = parseInt(process.env.OUTDATED_LOOMIES_TIMEOUT);

  // Remove if the generated_at field is older than generated_at + outdatedTimeout
  const outdatedLoomies = await WildLoomieModel.find({
    generated_at: { $lt: currentUnixTime - outdatedTimeout * 60 },
  });

  const loomiesIds = outdatedLoomies.map((loomie) => loomie._id);
  console.log("Loomies to remove: ", loomiesIds.length);

  await ZoneModel.updateMany({}, { $pullAll: { loomies: loomiesIds } });
  await WildLoomieModel.deleteMany({ _id: { $in: loomiesIds } });
  console.log("🔥 Done!");
}

// Show differences
const wildLoomiesAfter = await WildLoomieModel.find({}).count();
console.log(
  `\nInfo: ${wildLoomiesBefore - wildLoomiesAfter} wild loomies removed.`
);

// Disconnect from MongoDB
mongoose.disconnect();
