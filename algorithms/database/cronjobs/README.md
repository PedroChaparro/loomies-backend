# 💽 Database Cronjobs

This directory contains the cronjobs (scheduled tasks) related to the database.

## 🧹 Clear Loomies

This cronjob removes outdated loomies from the database.

### Outdated Loomies

```bash
pnpm clear_loomies:outdated
```

**Execution period:** Every 24 hours

### All Loomies

```bash
pnpm clear_loomies:all
```

**Note:** Before running the clear:outdated command, make sure you have a .env file in the `algorithms/database` directory with the OUTDATED_LOOMIES_TIMEOUT variable set to the desired value

## ✨ Gyms rewards

This cronjob updates the rewards of the gyms in the database.

- To update the rewards of the gyms run the following command from the `algorithms/database` directory:

```bash
pnpm update:rewards
```

**Execution period:** Every 24 hours
