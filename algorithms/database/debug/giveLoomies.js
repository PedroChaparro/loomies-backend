import dotenv from "dotenv";
import mongoose from "mongoose";

import { BaseLoomieModel, UserModel } from "../models/mongoose.js";

import { giveAllLoomies } from "../utils/utils.js";

// connect to MongoDB
dotenv.config();
mongoose.set("strictQuery", true);
mongoose.connect(process.env.MONGO_URI, { dbName: "loomies" });

// get owner id
const ownerId = process.argv[2];
if (!ownerId) {
  console.log("OwnerId required");
  process.exit();
}

// check owner exists
try {
  const owner = await UserModel.findById(ownerId);
  if (!owner) throw "Owner not found";
} catch (e) {
  console.error(e);
  process.exit();
}

// get base Loomies
const baseLoomies = await BaseLoomieModel.find();
console.log(`Found ${baseLoomies.length} base Loomies`);

// create all loomies
console.log(`Inserting to user ${ownerId}...`);
await giveAllLoomies(baseLoomies, ownerId);
console.log("Finished");

// close connection
mongoose.connection.close();
