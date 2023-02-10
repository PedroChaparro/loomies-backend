import dotenv from "dotenv";
import fs from "fs";
import mongoose from "mongoose";
dotenv.config();

// Connect to MongoDB
mongoose.connect(process.env.MONGO_URI, { dbName: "loomies" });

// Mongo models
const ZoneSchema = new mongoose.Schema(
  {
    leftFrontier: Number,
    rightFrontier: Number,
    topFrontier: Number,
    bottomFrontier: Number,
    number: Number,
    gym: { type: mongoose.Schema.Types.ObjectId, ref: "Gym" },
  },
  { versionKey: false }
);
const ZoneModel = mongoose.model("Zones", ZoneSchema);

const GymSchema = new mongoose.Schema(
  {
    latitude: Number,
    longitude: Number,
    name: String,
  },
  { versionKey: false }
);
const GymModel = mongoose.model("Gyms", GymSchema);

// Read data from json files
const zones = JSON.parse(fs.readFileSync("../../data/zones.json"));
const gyms = JSON.parse(fs.readFileSync("../../data/places.json"));

// Insert data into MongoDB
for await (const gym of gyms) {
  // Find zone from gym zone identifier
  const zoneIndex = zones.findIndex(
    (zone) => zone.identifier === gym.zoneIdentifier
  );
  const zone = zones[zoneIndex];

  // Insert gym into mongodb and get the id
  const { name, latitude, longitude } = gym;
  const newGym = new GymModel({ name, latitude, longitude });
  const { _id } = await newGym.save();

  // Insert zone with the gym id
  const { leftFrontier, rightFrontier, topFrontier, bottomFrontier, number } =
    zone;
  const newZone = new ZoneModel({
    leftFrontier,
    rightFrontier,
    topFrontier,
    bottomFrontier,
    number,
    gym: _id,
  });
  await newZone.save();

  // Remove zone from zones array
  zones.splice(zoneIndex, 1);
}

// Insert remaining zones without gym
console.log(`Inserting ${zones.length} zones without gym...`);
for await (const zone of zones) {
  const { leftFrontier, rightFrontier, topFrontier, bottomFrontier, number } =
    zone;

  const newZone = new ZoneModel({
    leftFrontier,
    rightFrontier,
    topFrontier,
    bottomFrontier,
    number,
  });

  await newZone.save();
}

// Close connection
mongoose.connection.close();
