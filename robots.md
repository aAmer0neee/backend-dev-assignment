# при имеющемя условии для задачи невозможно конечное решение

- роботы имеют одинаковую программу без возможности определения своей позиции относительно флага
- удинственный способ поиск флага - движение в одну из сторон,
что гарантированно приведет робота с противоположной стороны к флагу,
- а второго к движению в бесконечность,т.к не сможет определить когда развернутся
- отсутствует память для хранения количества пройденых шагов

## единственное решение - бесконечный цикл итераций

### 1: MR
### 2: IF FLAG
### 3: GOTO 3

### 4: ML
### 5: ML
### 6: IF FLAG
### 7: GOTO 3

### 8: MR
### 9: MR
### 10: MR
### 11: IF FLAG
### 12: GOTO 3
... ***до бесконечности

