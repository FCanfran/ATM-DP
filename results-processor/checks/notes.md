


# By fixed number of cores

- `results-fix-cores-main.sh`: runs for each of the specified core values labels, the program `results-fix-cores-a.sh`.

- `results-fix-cores-a.sh`: run of the `dieffpy-cores.py` for the specified core value experiment.


## `Dieffpy-cores.py` 

### Plots

#### `execTime.png`

#### `mrt.png`

#### `radar-diefk.png`

#### `radar-dieft.png`

Modified so to:

- Replace Checks by inverse mrt (mrt^-1) -> it is more informative, since
the number of checks it is expected to be the same in all the cases.

#### `traces-response-time-reduced.png`

- Transform the unit of the responseTime field values: from ns to ms (after custom load_trace - traces_response_time)

#### `traces-response-time.png`

- Transform the unit of the responseTime field values: from ns to ms (after custom load_trace - traces_response_time)


#### `traces.png`

### Results text file: `dieffpy-out.txt`

Modified so to:

- include checks (instead of labeling with comp)
- include mrt
- detail of the units of tfft (ms), mrt (ms) and totaltime (s).

- switch the diefk outputs: 
    - 5 -> 500
    - 10 -> 1000 first answers

TODO: For the ongoing 
- add the alerts values

# Usage

## By cores

Obtain processed results and plots by cores (for each of the indicated number of cores variations):
```
$> ./results-fix-cores-main.sh <results-directory> <testName>
```

## By filters 

```
$> ./results-fix-filters-main.sh <results-directory> <testName>
```

## Combined

```
$> ./results-combined.sh <resultsDirectoryPath> <TEST(name)> <DO_JOIN(0:no,1:yes)> <num_interactions>"
```

- `num_interactions`: refers to the number of interactions (openings and closings of transactions), therefore `num_tx` x 2.



%%%%%%%
% Stream Size - Bank Size configurations
%
@misc{exps-atmfrauddetectionstreamdata,
      title={ATM Fraud Detection using Streaming Data Analytics}, 
      author={Yelleti Vivek and Vadlamani Ravi and Abhay Anand Mane and Laveti Ramesh Naidu},
      year={2023},
      eprint={2303.04946},
      archivePrefix={arXiv},
      primaryClass={cs.LG},
      url={https://arxiv.org/abs/2303.04946}, 
}

@article{exps-costsensitivepayment,
author = {Nami, Sanaz and Shajari, Mehdi},
year = {2018},
month = {06},
pages = {},
title = {Cost-sensitive payment card fraud detection based on dynamic random forest and k-nearest neighbors},
volume = {110},
journal = {Expert Systems with Applications},
doi = {10.1016/j.eswa.2018.06.011}
}
@ARTICLE{exps-creditcardfrauddetectionusinghmm,
  author={Srivastava, Abhinav and Kundu, Amlan and Sural, Shamik and Majumdar, Arun},
  journal={IEEE Transactions on Dependable and Secure Computing}, 
  title={Credit Card Fraud Detection Using Hidden Markov Model}, 
  year={2008},
  volume={5},
  number={1},
  pages={37-48},
  keywords={Credit cards;Hidden Markov models;State estimation;Neural networks;Electronic commerce;Security;Internet telephony;Information analysis;Pattern analysis;Electronic Commerce;Security and Protection;Electronic Commerce;Security and Protection},
  doi={10.1109/TDSC.2007.70228}
}
@article{exps-featureengineering,
title = {Feature engineering strategies for credit card fraud detection},
journal = {Expert Systems with Applications},
volume = {51},
pages = {134-142},
year = {2016},
issn = {0957-4174},
doi = {https://doi.org/10.1016/j.eswa.2015.12.030},
url = {https://www.sciencedirect.com/science/article/pii/S0957417415008386},
author = {Alejandro {Correa Bahnsen} and Djamila Aouada and Aleksandar Stojanovic and Björn Ottersten},
keywords = {Cost-sensitive learning, Fraud detection, Preprocessing, Von Mises distribution},
abstract = {Every year billions of Euros are lost worldwide due to credit card fraud. Thus, forcing financial institutions to continuously improve their fraud detection systems. In recent years, several studies have proposed the use of machine learning and data mining techniques to address this problem. However, most studies used some sort of misclassification measure to evaluate the different solutions, and do not take into account the actual financial costs associated with the fraud detection process. Moreover, when constructing a credit card fraud detection model, it is very important how to extract the right features from the transactional data. This is usually done by aggregating the transactions in order to observe the spending behavioral patterns of the customers. In this paper we expand the transaction aggregation strategy, and propose to create a new set of features based on analyzing the periodic behavior of the time of a transaction using the von Mises distribution. Then, using a real credit card fraud dataset provided by a large European card processing company, we compare state-of-the-art credit card fraud detection models, and evaluate how the different sets of features have an impact on the results. By including the proposed periodic features into the methods, the results show an average increase in savings of 13%.}
}
