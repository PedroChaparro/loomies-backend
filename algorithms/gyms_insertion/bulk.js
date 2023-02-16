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
    coordinates: String,
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
const coordinates = { x: 0, y: 0 };
let currentLongitude;

console.log("Expected zones: ", zones.length);
console.log("Expected gyms: ", gyms.length);

for await (const zone of zones) {
  let GymMongoId;

  // Initialize currentLongitude
  if (!currentLongitude) currentLongitude = zone.bottomFrontier;

  // Increment coordinates when longitude changes (New row)
  if (currentLongitude !== zone.bottomFrontier) {
    currentLongitude = zone.bottomFrontier;
    coordinates.x = 0;
    coordinates.y++;
  }

  // Get the zone's gym
  const gym = gyms.findIndex((gym) => gym.zoneIdentifier === zone.identifier);

  // Insert the gym into mongodb and get the id
  if (gym !== -1) {
    const { name, latitude, longitude } = gyms[gym];
    const newGym = new GymModel({ name, latitude, longitude });
    const { _id } = await newGym.save();
    GymMongoId = _id;
  }

  // Insert zone with the gym id
  const { leftFrontier, rightFrontier, topFrontier, bottomFrontier, number } =
    zone;

  const newZone = new ZoneModel({
    leftFrontier,
    rightFrontier,
    topFrontier,
    bottomFrontier,
    number,
    coordinates: `${coordinates.x},${coordinates.y}`,
    gym: GymMongoId ? GymMongoId : null,
  });

  await newZone.save();

  // Increment coordinates
  coordinates.x++;
}

console.log("Zones inserted: ", await ZoneModel.countDocuments());
console.log("Gyms inserted: ", await GymModel.countDocuments());

// Close connection
mongoose.connection.close();
