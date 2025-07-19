graph = {
genesOfInterest: [
"differential expression",
],
nodes: [
{id:"FRS2",samples: ["QA==",]},
{id:"PIK3CD",samples: ["QA==",]},
{id:"SRC",samples: ["",]},
{id:"RALA",samples: ["",]},
{id:"KRAS",samples: ["",]},
{id:"PLCE1",samples: ["",]},
{id:"PIK3C3",samples: ["",]},
],
links: [
{source:"KRAS", target:"RALA", type:"pp"},
{source:"PIK3C3", target:"PLCE1", type:"pp"},
{source:"PLCE1", target:"KRAS", type:"pp"},
{source:"FRS2", target:"PIK3C3", type:"pp"},
{source:"PIK3CD", target:"PLCE1", type:"pp"},
{source:"PLCE1", target:"PIK3C3", type:"pp"},
{source:"FRS2", target:"PIK3CD", type:"pp"},
{source:"FRS2", target:"SRC", type:"pp"},
{source:"PLCE1", target:"PIK3CD", type:"pp"},
{source:"PIK3CD", target:"FRS2", type:"pp"},
{source:"SRC", target:"RALA", type:"pp"},
],
conditions: [
"sample1",
"sample2",
],
}