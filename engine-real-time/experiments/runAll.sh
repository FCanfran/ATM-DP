#!/bin/bash

job1=$(sbatch --parsable ./launchAll-1c.sh exps-descriptions/small/1c 10)

job2=$(sbatch --parsable --dependency=afterok:$job1 ./launchAll-2c.sh exps-descriptions/small/2c 10)

job3=$(sbatch --parsable --dependency=afterok:$job2 ./launchAll-4c.sh exps-descriptions/small/4c 10)

job4=$(sbatch --parsable --dependency=afterok:$job3 ./launchAll-8c.sh exps-descriptions/small/8c 10)

sbatch --dependency=afterok:$job4 ./launchAll-16c.sh exps-descriptions/small/16c 10
