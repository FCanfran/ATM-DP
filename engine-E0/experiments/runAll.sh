#!/bin/bash

job1=$(sbatch --parsable ./launchAll-1c.sh exps-descriptions/small/1c 2)

job2=$(sbatch --parsable --dependency=afterok:$job1 ./launchAll-4c.sh exps-descriptions/small/4c 2)

sbatch --dependency=afterok:$job2 ./launchAll-16c.sh exps-descriptions/small/16c 2
