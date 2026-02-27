# GoNetic 2

GoNetic 2 is a tool to identify subnetworks of interest in a gene interaction network using mutation and expression data. It applies NSGA-II optimization to select subnetworks that are both relevant to the omics data and consistent with the network structure.

GoNetic 2 is a complete rewrite of the original tool. It introduces support for expression data and a revised subnetwork selection procedure. The core idea remains unchanged: omics data is mapped to a network; paths of interest are identified and compiled into a logical formula; subnetworks are then evaluated for the presence of these paths using that formula.

The original GoNetic version (v1.0.0) is still available under tag `v1.0.0` in this repository.

This repository contains the source code, binaries are provided in the Releases section.

The binaries and code in this repository are only licensed for non-commercial, academic, and educational use.

### requirements and setup
One of the following is required to run GoNetic:
 - use the provided binaries from a release version
 - Go 1.23+ to build the source code, see [Go installation instructions](https://go.dev/doc/install) for details on how to install Go.

The following additional requirements are needed to run GoNetic:
 - c2d [1] or d4 [2] compiler:
   - c2d binaries can be acquired for Linux or Windows here: <http://reasoning.cs.ucla.edu/c2d/>, place it in the `etc/` directory under the name `c2d` or `c2d.exe`, and set execution permissions. 
     - You might have to install a compatibility library to run the c2d binary on your system, since it is a 32-bit binary.
     - If you do not rename the c2d binary, you can instruct GoNetic to use this binary by setting: `--ddnnf-compiler /path/to/c2dbinary --ddnnf-type c2d` 
   - d4 source code can be acquired from [github.com/crillab/d4](https://github.com/crillab/d4), see build instructions further down in this readme
 - WFG-hypervolume [3]:
   - wfg-hypervolume source code can be acquired from [github.com/lbradstreet/WFG-hypervolume](https://github.com/lbradstreet/WFG-hypervolume), see build instructions further down in this readme

#### Building GoNetic

1. Download and install Go 1.23+ from [go.dev/dl/](https://go.dev/dl/).
2. Clone the GoNetic repository from [github.com/MarchalLab/gonetic](https://github.com/MarchalLab/gonetic).
3. Build the GoNetic executable by running `go build` in the `gonetic` folder. This creates the GoNetic binary in the `gonetic` folder.

#### Building d4

The d4 compiler must be built using a GCC-compatible C++ compiler (e.g., GCC or Clang) and `make`. On Windows, it is recommended to use WSL, or alternatively, use a precompiled `c2d` binary instead.

* On Linux and WSL, ensure that `make` and a GCC-compatible compiler are installed (e.g., via `build-essential` on Ubuntu).
* On macOS, install Xcode Command Line Tools by running `xcode-select --install`.

1. Clone the d4 repository from [github.com/crillab/d4](https://github.com/crillab/d4).
2. Build the executable by running `make` in the `d4` directory. This produces the `d4` binary.
3. Place the generated d4 executable in the `etc` folder.
4. Alternatively, instruct GoNetic to use this binary by setting: `--ddnnf-compiler /path/to/d4binary --ddnnf-type d4`

#### Building WFG-hypervolume

The WFG hypervolume implementation must be built using a GCC-compatible compiler (e.g., GCC or Clang) and `make`. On Windows, use MSYS2 or a UNIX-like environment such as WSL. The provided Makefile is not compatible with MSVC.

* On Linux and WSL, ensure that `make` and a GCC-compatible compiler are installed (e.g., via `build-essential` on Ubuntu).
* On macOS, install Xcode Command Line Tools by running `xcode-select --install`.
* On Windows using MSYS2, open the **MinGW64 shell** and install the required tools `pacman -S mingw-w64-x86_64-gcc make`

1. Clone the WFG-hypervolume repository from [github.com/lbradstreet/WFG-hypervolume](https://github.com/lbradstreet/WFG-hypervolume).
2. Copy the file `etc/avl.h` into the WFG-hypervolume directory.
3. Build the WFG executables by running `make march=native`. This produces `wfg0`, `wfg1`, and `wfg2` (or `.exe` files on Windows).
4. Place the generated WFG executables in the `etc` folder.

### usage
GoNetic has three subcommands:

- `./gonetic QTL -h`: GoNetic looks for paths between mutations across samples.
- `./gonetic EQTL -h`: GoNetic looks for paths between mutations across samples, and for paths from mutations to differentially expressed genes within samples.
- `./gonetic expression -h`: GoNetic looks for paths between differentially expressed genes within samples.

All three subcommands are illustrated in the script `run-example.sh`, using the data in `example-data`.

### file formats
The mutations file is a tab or comma separated file with a header line that starts with a `#`-character. The following columns are required:
 - `gene name`: an identifier of the mutated gene, should match the identifier of that gene in the network file
 - `condition`: an identifier of a sample or condition
The following columns are optional:
 - `functional score`: an impact score between 0 and 1, e.g. CADD scores. If this column is not present or should be ignored, add `-e=false` to the command.
 - `freq increase`: a frequency score between 0 and 1, e.g. variant allele frequency. If this column is not present or should be ignored, add `-c=false` to the command.
Additional columns can be present in the file, but are ignored by GoNetic.

The network file is a tab or comma separated file with a header line for each type of interaction that occurs in the network.
Header lines are of the form `% <interaction identifier> [non-]regulatory`.
Interaction entries have 5 columns: 
 - source gene name
 - sink gene name
 - interaction type identifier (e.g. pp for protein-protein interactions)
 - "directed" for directed interactions, or "undirected" for bidirectional interactions
 - an edge weight between 0 and 1

Example files can be found in the `example-data` folder, here we show the header and the first 2 entries of these files.
- expression.csv
```
#gene name,lfc,pval,condition
gene_530,-1.03334259198456e-06,0.211525759416019,sample_4
```

- mutations.csv
```
#gene name,reference,alternative,chrom,start,end,freq increase,functional score,condition
gene_624,C,-,1,953261,953261,0.0779220779220779,1.0,sample_1
```

- network.csv
```
% pp non-regulatory
gene_853,gene_123,pp,directed,0.84
```

- targets.csv
```
#gene name,lfc,pval,condition
gene_750,3.50871582582474,0.000154767761237015,sample_4
```

After running the script `run-example.sh` the output can be found in the `example-result/` folder, where there is one subdirectory per command. In these subdirectories the main result files can be found in `resulting_networks/normsum`:
 - `d3js_visualization`: a html+js visualisation of the resulting network.
 - `weighted.network`: a tab separated file containing the resulting network.
 - `conditionSpecificMutationRanking.txt`: a tab separated file containing all genes that are in the resulting network that are also mutated in the input data. The genes are ranked based on the cumulative score of paths in this subnetwork that start in this gene.
 - `conditionSpecificDeRanking.txt`: a tab separated file containing all genes that are in the resulting network that are also differentially expressed in the input data. The genes are ranked based on the cumulative score of paths in this subnetwork that start in this gene.


### visualization
#### gene sets
Gene sets can be visualized in the d3js_visualization as follows:
1. Create a new JavaScript file named `genesets.js`.

2. In this file define exactly one top level variable named `geneSets`.
   Use the pattern:

```{js}
const geneSets = {
    SET_NAME_1: [
        'GENE_A', 'GENE_B'
    ],
    'Custom set 2': [
        'GENE_C', 'GENE_D'
    ]
};
```

3. Each property of `geneSets` is one gene set.
   - Key: the gene set name, written as either an identifier or a quoted string.
   - Value: an array of gene symbols written as strings. These arrays are flat lists without nesting.

4. Add additional gene sets by adding more properties to the object.

5. When finished, place `genesets.js` in the `d3js_visualization` folder next to the existing files so that the visualization code in `highestScoringSubnetwork.html` can access the `geneSets` variable.


### references
[1] Darwiche A. New advances in compiling CNF to decomposable negation normal form. Proc. of ECAI, 328-332.  
[2] Lagniez J-M, Marquis P. On Preprocessing Techniques and Their Impact on Propositional Model Counting in Journal of Automated Reasoning (JAR), vol. 58, n° 4, pp. 413-481, 2017.  
[3] While L, Bradstreet L, and Barone L. A Fast Way of Calculating Exact Hypervolumes. IEEE Transactions on Evolutionary Computation 16(1), 2012
[4] Jassal B, Matthews L, Viteri G, Gong C, Lorente P, Fabregat A, Sidiropoulos K, Cook J, Gillespie M, Haw R, Loney F, May B, Milacic M, Rothfels K, Sevilla C, Shamovsky V, Shorser S, Varusai T, Weiser J, Wu G, Stein L, Hermjakob H, D'Eustachio P. The reactome pathway knowledgebase. Nucleic Acids Res. 2020 Jan 8;48(D1):D498-D503. doi: 10.1093/nar/gkz1031. PubMed PMID: 31691815.  
