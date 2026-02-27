dir=example-data
out=example-result

rm -fr ./${out}
mkdir ./${out}

# to reduce run time for this example, we use a target network size of 25 (default 100) and a population size of 100 (default 500)
settings="-q etc -n ${dir}/network.csv --path-length 3 --target-network-size 25 --population-size 100"

# run QTL mode on sample data
./gonetic QTL ${settings} -d ${dir}/mutations.csv -o ${out}/1.qtl

# run EQTL mode on sample data
./gonetic EQTL ${settings} -d ${dir}/mutations.csv -a ${dir}/expression.csv -g ${dir}/targets.csv -o ${out}/2.eqtl

# run expression mode on sample data
./gonetic expression ${settings} -q etc -n ${dir}/network.csv -a ${dir}/expression.csv -g ${dir}/targets.csv -o ${out}/3.expression
