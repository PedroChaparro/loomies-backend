# Gyms insertion

**Make sure you've generated the `places.json` and `zones.json` files inside the `data/` folder before running the bulk script**

## Instructions to setup the database

1. Install packages:

```bash
pnpm i
# npm i
```

2. Create `.env` file with the `MONGO_URI`.

3. Run `bulk` script:

```bash
pnpm bulk
# npm run bulk
```

4. Run `test` script to ensure the insertion was correct:

```bash
pnpm test
# npm run test
```
