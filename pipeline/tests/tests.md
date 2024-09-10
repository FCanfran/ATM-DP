# Testing

Tests regarding the following aspects are performed. For them we use the tx gathered in the edited 
`transaction-tests.csv` file:

```
transaction_id,number_id,ATM_id,transaction_start,transaction_end,transaction_amount
0,c-NIGER-0,OGUN-0,2018-04-01 02:40:23,2018-04-01 02:49:03,52537.4
1,c-NIGER-0,BENUE-4,2018-04-01 16:02:09,2018-04-01 16:05:49,29201.12
2,c-NIGER-0,OGUN-0,2018-04-02 10:30:27,2018-04-02 10:36:58,37940.13
7,c-NIGER-1,OGUN-1,2018-04-03 02:39:36,2018-04-03 02:45:32,43667.03
9,c-NIGER-2,OGUN-1,2018-04-03 10:23:07,2018-04-03 10:31:49,17697.11
10,c-NIGER-2,OGUN-1,2018-04-03 15:16:17,2018-04-03 15:21:25,20806.12
3,c-NIGER-0,BENUE-4,2018-04-03 22:50:14,2018-04-03 22:56:02,24739.97
4,c-NIGER-0,BENUE-4,2018-04-04 06:58:51,2018-04-04 07:06:11,8700.09
5,c-NIGER-0,BENUE-4,2018-04-04 14:11:42,2018-04-04 14:16:38,23634.22
11,c-NIGER-2,OGUN-1,2018-04-05 08:53:14,2018-04-05 08:58:21,33732.43
6,c-NIGER-0,OGUN-0,2018-04-05 10:38:47,2018-04-05 10:44:12,29682.4
8,c-NIGER-1,OGUN-3,2018-04-05 12:54:18,2018-04-05 12:59:16,16966.09
12,c-NIGER-2,OGUN-1,2018-04-05 16:33:38,2018-04-05 16:38:01,22325.85
13,c-NIGER-2,OGUN-2,2018-04-05 22:33:50,2018-04-05 22:40:24,57999.49
```

It gathers transactions from 3 different cards in a time period of 5 days, from 2018-04-01 to 2018-04-05.


### Volatile subgraph

Objectives:
- Addition: adds new tx to the corresponding filter volatile subgraph correctly
- Update: taking the timestamp of the tx of the pipeline (we assume they are ordered 
in time: *property graph stream* definition 5.2 of Seraph article), we update the filter
subgraph in order to only have it in a state corresponding to a certain **window** of time
(e.g. the last `timeTransactionThreshold`) -> so that no *outdated* edges/tx are no longer stored, and therefore they are deleted from the filter subgraph.

### 1. Addition

To simply check this and avoid other operations on the filters being performed, we set:

- `timeTransactionThreshold` = 10 * 24 * time.Hour // 10 days
- `timeFilterThreshold` = 10 * 24 * time.Hour      // 10 days 

so that all the edges of the filters' volatile subgraphs are kept and no filters are destroyed.
The objective is to see that all the tx are added to the corresponding filter subgraph.

Result in `test-add.txt`:

*Note that we only print the tx ids to show the edges that are in each of the subgraphs*

- Subgraph c-NIGER-0: 0-1-2-3-4-5-6
- Subgraph c-NIGER-1: 7-8
- Subgraph c-NIGER-2: 9-10-11-12-13

### 2. Update

Check that the subgraphs of the filters are updated accordingly to the `timeTransactionThreshold` defined time and avoiding the deletion of filters so that
we only focus in the update of the subgraphs. 

**Note**: remember the problem of the temporal outdate of some of the filters. That is,
due to the nature of the pipeline so far, the timestamps only reach those filters through which the corresponding tx reaches to flow, not reaching those after the corresponding filter to which that tx belongs.

- `timeTransactionThreshold` = 1 * 24 * time.Hour  // 1 days
- `timeFilterThreshold` = 10 * 24 * time.Hour      // 10 days 

Result in `text-update.txt`:

The sequence of what it is expected to happen is shown in the following table:

| filter 0 | filter 1 | filter 2 |
|----------|----------|----------|
| 0        |          |          |
| 0,1      |          |          |
| 1,2      |          |          |
| 2        | 7        |          |
| 2        | 7        | 9        |
|          | 7        | 9,10     |
| 3        | 7        | 9,10     |
| 3,4      | 7        | 9,10     |
| 3,4,5    | 7        | 9,10     |
| 5        |          | 11       |
| 5,6      |          | 11       |
| 5,6      | 8        | 11       |
| 6        | 8        | 11,12    |
| 6        | 8        | 11,12,13 |


TODO: Gestionar cómo hacer cuando tras un update no queda ningún edge en el subgraph...
(eliminar filtro, mantener?) --> Esto no puede ocurrir tal y como lo tenemos implementado
--> ver explicación en el cuadernillo!!!  ----- SOLUCIONADO... (Si que puede haber filters que 
tengan su subgraph vacío, gracias a variable `last_timestamp` se gestiona, eliminando el filter una
vez se haya detectado que se ha sobrepasado el `timeFilterThreshold` time.) - explicado en el overleaf
y en el cuadernillo (opc 2).


### Filter lifetime management

Objectives:
- Delete/Destroy the corresponding filters that do not register any activity (no new tx related to them) in the last `timeFilterThreshold` time. 
- Check correct pipeline reconnection


*Test file:* `transaction-tests-kill.csv`.

| filter 0 | filter 1 | filter 2 |
|----------|----------|----------|
| 0        |          |          |
| 0,1      |          |          |
| 1,2      |          |          |
| 2        | 7        |          |
| 2        | 7        | 9        |
|          | 7        | 9,10     |
| 3        | 7        | 9,10     |
| 3,4      | 7        | 9,10     |
| 3,4,5    | 7        | 9,10     |
| 5        | x        | 11       |
| filter 0 | filter 2 |          |
| 5,6      | 11       |          |
| 6        | 11,12    |          |
| 6        | 11,12,13 |          |
| filter 0 | filter 2 | filter 1 |
| 6        | 11,12,13 | 8        |
|          | 12,13    | 8,14     |
| x        | 15       | 8,14     |
| filter 2 | filter 1 |          |
| 15       | 16       |          |


Qué probamos:

- Subgrafo vacío pero filtro no se destruye -> causante tx 10, filtro 0
- Filtro destruido directamente (sin subgrafo vacío) -> causa tx 11, filtro 1
- Filtro destruido tras haber pasado fase de tener subgrafo vacío -> causa tx 15, filtro 0

Vida filtros y orden:
- 0-1-2
- 0-2
- 0-2-1
- 2-1