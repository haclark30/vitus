Todos

- [x] update fitbit to use subcommands
- [x] command to just update single day of weight in db
    - [ ] clean up commands for loading all data vs just partial amount
- [ ] start with bubbletea
- [ ] design db
- [ ] better error handling

---
idea:
- depending on which data is requested, cli first checks db for most recent entry
- if it's older than some amount of time, call api and update data
- if it's not, just read from db
- have a table just for that, most recent successful sync


notes:
- store minutes data as ranges, so if i have 0 steps in several hours straight, store the start and end minute to save space
