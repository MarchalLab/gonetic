# GoNetic binary
This repository contains a static binary of GoNetic for 64-bit linux.
The content of this repository is only licensed for non-commercial academic use.

### requirements and setup
 - 64-bit linux

 - c2d compiler [1]:
   - We are in no way affiliated with the c2d project. See the c2d manual for licensing information of the c2d compiler: "The c2d compiler is licensed only for non–commercial, research and educational use."
   - Get the linux binary `c2d_linux` here: <http://reasoning.cs.ucla.edu/c2d/>, place it in the `etc/` directory, and set execution permissions (e.g. `chmod u+x c2d_linux`) 
   - Install `libc6:i386` because the c2d binary is a 32-bit executable

### usage
`./gonetic QTL -h`

### example
`./gonetic QTL -q etc -n sample/network.txt -d sample/mutations.csv -o output`
Other parameters are optional.

### file formats
The mutations file is a tab or comma separated file with a headerline that starts with a `#`-character. The following columns are required:
 - `gene name`: an identifier of the mutated gene, should match the identifier of that gene in the network file
 - `condition`: an identifier of a sample or condition
The following columns are optional:
 - `functional score`: an impact score between 0 and 1, e.g. CADD scores. If this column is not present or should be ignored, add `-e=false` to the command.
 - `freq increase`: a frequency score between 0 and 1, e.g. variant allele frequency. If this column is not present or should be ignored, add `-a=false` to the command.
Additional columns can be present in the file, but are ignored by GoNetic.

The network file is a tab or comma separated file with a header line for each type of interaction that occurs in the network.
Header lines are of the form `% <interaction identifier> [non-]regulatory`.
Interaction entries have 5 columns: 
 - source gene name
 - sink gene name
 - interaction type identifier (e.g. pp for protein-protein interactions)
 - "directed" for directed interactions, or "undirected" for bidirectional interactions
 - an edge weight between 0 and 1

Example files can be found in the `sample` folder, here we show the header and the first 2 entries of these files.
 - mutations.csv:
```
#gene name,condition,functional score,freq increase
PRDM16,Ls420,0.810177877122851,0.581176532205082
WRAP73,NYU160,0.816136367036238,0.863155362397098
```
 - network.txt [2]:
```
% pp non-regulatory
% pd regulatory
A2M,APOA1,pp,directed,1.0
A2M,BMP1,pp,directed,1.0
```

The main output files are in the `output/resulting_networks` folder:
 - `d3js_visualization`: a html+js visualisation of the resulting network, tested in Firefox and Chromium-based browsers.
 - `weighted.network`: a tab separated file containing the resulting network. The same type of header lines as in the input network file, each entry now consists of two columns: (1) an unweighted interaction in the same format as the input network file, and (2) the highest edge penalty for which this interaction was selected in the subnetwork selection phase.
 - `rankedMutations.txt`: a tab separated file containing all genes that are in the resulting network that are also mutated in the input data. The rank of the gene is based on the highest edge penalty for which this gene was selected in the subnetwork selection phase, where rank "1" corresponds with the highest edge penalty that lead to a valid subnetwork.

### references
[1] Darwiche A. New advances in compiling CNF to decomposable negation normal form. Proc. of ECAI, 328-332  
[2] Jassal B, Matthews L, Viteri G, Gong C, Lorente P, Fabregat A, Sidiropoulos K, Cook J, Gillespie M, Haw R, Loney F, May B, Milacic M, Rothfels K, Sevilla C, Shamovsky V, Shorser S, Varusai T, Weiser J, Wu G, Stein L, Hermjakob H, D'Eustachio P. The reactome pathway knowledgebase. Nucleic Acids Res. 2020 Jan 8;48(D1):D498-D503. doi: 10.1093/nar/gkz1031. PubMed PMID: 31691815.  
