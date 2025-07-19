graph = {
genesOfInterest: [
"differential expression",
],
nodes: [
{id:"SRC",samples: ["",]},
{id:"KRAS",samples: ["",]},
{id:"RALA",samples: ["",]},
{id:"PIK3CD",samples: ["QA==",]},
{id:"FRS2",samples: ["QA==",]},
{id:"PIK3C3",samples: ["",]},
{id:"PLCE1",samples: ["",]},
],
links: [
{source:"FRS2", target:"PIK3CD", type:"pp"},
{source:"FRS2", target:"SRC", type:"pp"},
{source:"PLCE1", target:"KRAS", type:"pp"},
{source:"FRS2", target:"PIK3C3", type:"pp"},
{source:"SRC", target:"RALA", type:"pp"},
{source:"KRAS", target:"RALA", type:"pp"},
{source:"PLCE1", target:"PIK3CD", type:"pp"},
{source:"PIK3CD", target:"PLCE1", type:"pp"},
{source:"PIK3CD", target:"FRS2", type:"pp"},
{source:"PIK3C3", target:"PLCE1", type:"pp"},
{source:"PLCE1", target:"PIK3C3", type:"pp"},
],
conditions: [
"sample1",
"sample2",
],
}