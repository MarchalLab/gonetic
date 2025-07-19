class GoNetic {
    constructor() {
    };
    init(graph, paths, geneSets, radius) {
        // keep a copy of graph argument
        this.graph = graph;
        this.paths = paths;
        this.parsePaths();
        this.geneSets = geneSets;
        // size settings
        this.radius = radius;
        this.scalingFactor = 1.6;
        this.strokeWidth = 6 * this.scalingFactor * 1.3;
        this.fontSize = 24 * this.scalingFactor;
        this.pieScalingFactor = 1;
        this.pieSize = this.scalingFactor * this.pieScalingFactor;
        this.basePieSize = .375 * this.pieSize;
        this.pieBorderSize = .125 * this.pieSize;
        this.pieSizeIncrease = .5625 * this.pieSize;
        this.nodeSize = this.radius * this.scalingFactor;
        this.pieLimit = 24;
        // set highlight mode, neighbors, component, or paths
        this.mouseoverEnabled = true;
        this.highlightMode = 'paths';
        // add buttons
        this.addHighlightButtons();
        this.addGeneSetButtons();
        this.addSampleButtons();
        // add info
        this.infoElement = document.getElementById('infoText');
        this.defaultInfo = 'Hover over a node to display available information.';
        this.setInfo(this.defaultInfo);
        // create styles
        this.createStyleSheet();
        // determine svg size based on number of nodes
        this.nodeCount = this.graph.nodes.length;
        this.width = 300 * Math.sqrt(this.nodeCount);
        this.height = 200 * Math.sqrt(this.nodeCount);
        // gather genes of interest information
        this.genesOfInterest = [];
        this.graph.nodes.forEach(n => {
            for (const id in n.samples) {
                n.samples[id] = base64ToBools(n.samples[id], graph.conditions.length);
                if (this.genesOfInterest.includes(id)) {
                    continue;
                }
                this.genesOfInterest.push(id);
            }
        });
        // Preprocess: create link -> path data map for fast lookup and compute link weights
        this.edgeIndex = {};
        this.graph.links.forEach(link => {
            const key = `${link.source};${link.target}`;
            link.paths = this.pathsByEdge[key] || [];
            this.edgeIndex[key] = link;
            // compute link weight
            link.weight = 0;
            for (const path of this.pathsByEdge[key]) {
                link.weight += path.score;
            }
        });
        // compute node opacity
        this.graph.nodes.forEach(n => {
            n.opacity = 0.3;
            n.count = 0;
            for (const id in n.samples) {
                const count = n.samples[id].filter(c => c).length;
                if (count > 0) {
                    n.opacity = 1;
                }
                n.count += count;
            }
        });
        // count per geneofinterest
        this.graph.nodes.forEach(n => {
            n.countPerTrait = [];
            for (const id in n.samples) {
                const count = n.samples[id].filter(c => c).length;
                n.countPerTrait.push(count);
            }
        });
        // create geneSetMap
        this.geneSetMap = {};
        for (const setID in this.geneSets) {
            this.geneSetMap[setID] = {};
            for (const gene of this.geneSets[setID]) {
                this.geneSetMap[setID][gene] = true;
            }
        }
        // attach geneSet information to nodes
        this.graph.nodes.forEach(n => {
            n.geneSets = {};
            for (const setID in this.geneSets) {
                if (this.geneSetMap[setID][n.id]) {
                    n.geneSets[setID] = true;
                }
            }
        });
        // scale node size if not drawing pies
        this.graph.nodes.forEach(n => {
            let nodeRadius = this.nodeSize;
            if (!this.isDrawingPies()) {
                nodeRadius += n.count * this.pieSize
            }
            n.radius = nodeRadius;
        });
        // node size per trait
        this.graph.nodes.forEach(n => {
            n.radiusPerTrait = [];
            Object.values(n.countPerTrait).forEach(c => {
                n.radiusPerTrait.push(this.nodeSize + Math.sqrt(c) * this.pieSize);
            });
        });
        // attach labels to nodes
        const label = {
            nodes: [],
            links: [],
        };
        this.graph.nodes.forEach((d, i) => {
            label.nodes.push({node: d});
            label.nodes.push({node: d});
            label.links.push({
                source: i * 2,
                target: i * 2 + 1
            });
        });
        this.label = label;
        this.labelLayout = d3.forceSimulation(this.label.nodes)
            .force("charge", d3.forceManyBody().strength(-50))
            .force("link", d3.forceLink(this.label.links).distance(0).strength(2));
        // gather nodes and links from graph data
        this.nodeInDegrees = {};
        this.nodeOutDegrees = {};
        this.neighbors = {};
        this.maxLinkWeight = this.graph.links.reduce((mw, link) => Math.max(mw, link.weight), 0);
        this.graph.links.forEach(link => {
            link.opacity = (link.weight / this.maxLinkWeight * 7 + 1) / 8
            // increment node degrees
            this.nodeInDegrees[link.target] = (this.nodeInDegrees[link.target] | 0) + 1;
            this.nodeOutDegrees[link.source] = (this.nodeOutDegrees[link.source] | 0) + 1;
            if (this.neighbors[link.source] === undefined) {
                this.neighbors[link.source] = [];
            }
            this.neighbors[link.source].push(link.target);
            if (this.neighbors[link.target] === undefined) {
                this.neighbors[link.target] = [];
            }
            this.neighbors[link.target].push(link.source);
        });
        function traverse(node, group, neighbors, nodesById) {
            if ("group" in node) {
                node.group = Math.min(node.group, group);
            } else {
                node.group = group;
                neighbors[node.id].forEach(id => traverse(nodesById[id], group, neighbors, nodesById));
            }
        }
        this.nodesById = {};
        this.graph.nodes.forEach(node => this.nodesById[node.id] = node);
        this.graph.nodes.forEach((node, i) => traverse(node, i, this.neighbors, this.nodesById));
        // determine all the groups and assign them an index
        this.groups = {}
        let groupIdx = 0;
        this.graph.nodes.forEach(node => {
            if (this.groups[node.group] === undefined) {
                this.groups[node.group] = {
                    idx: groupIdx
                }
                groupIdx++;
            }
        })
        // set node group to group index, in this manner the groups are [0, ..., n]
        // set initial node positions, so clusters don't separate eachother
        // note the randomness of the position: if nodes have the same start position, it won't work
        this.graph.nodes.forEach(node => {
            node.group = this.groups[node.group].idx;
            node.x = this.width / groupIdx * (node.group + 0.5 + Math.random());
            node.y = this.height / groupIdx * (node.group + 0.5 + Math.random());
        });
        // create graph layout
        this.graphLayout = d3.forceSimulation(this.graph.nodes)
            .force("charge", d3.forceManyBody().strength(-3000))
            .force("center", d3.forceCenter(this.width / 2, this.height / 2))
            .force("x", d3.forceX(this.width).strength(0.2))
            .force("y", d3.forceY(this.height).strength(0.3))
            .force("cluster", forceCluster())
            .force("collide", forceCollide())
            .force("link", d3.forceLink(this.graph.links)
                .id(d => d.id)
                .distance(50)
                .strength(l => 3 * linkForceStrength(l))
            )
            .on("tick", ticked);
        // create adjacency list
        const adjList = {};
        this.graph.links.forEach(function (d) {
            adjList[d.source.index + "-" + d.target.index] = true;
            adjList[d.target.index + "-" + d.source.index] = true;
        });
        this.adjList = adjList;
        this.svg = d3.select("#svg-container")
            .append("svg")
            // Responsive SVG needs these 2 attributes and no width and height attr.
            .attr("preserveAspectRatio", "xMinYMin slice")
            .attr("viewBox", `0 0 ${this.width} ${this.height}`)
            // Class to make it responsive.
            .classed("svg-content-responsive", true);
        this.appendRule(
            '.svg-container',
            {
                height: '100%',
                width: '100%',
                'min-height': '100%',
            }
        );
        this.appendRule('.svg-content-responsive', {
            '-webkit-box-sizing': 'border-box',
            '-moz-box-sizing': 'border-box',
            'box-sizing': 'border-box',
            height: '100%',
            width: '100%',
            'max-height': '80vh',
            border: '1px',
            'border-color': 'black',
            'border-style': 'solid',
        });
        this.container = this.svg.append("g");
        this.svg.call(
            d3.zoom()
                .scaleExtent([.1, 10])
                .on("zoom", (event) => this.container.attr("transform", event.transform))
        );
        this.link = this.container.append("g").attr("class", "links")
            .selectAll("path")
            .data(this.graph.links)
            .enter()
            .append("svg:path")
            .attr("class", function (d) {
                return "link " + d.type;
            })
            .attr("marker-end", function (d) {
                return "url(#" + d.type + ")";
            })
            .attr("marker-start", function (d) {
                if (d.direction === "undirected") {
                    return "url(#" + d.type + "undir" + ")";
                }
            })
            .attr("stroke", "#aaa")
            .attr("stroke-width", d => 5 + d.opacity * 12)
            .attr("opacity", d => d.opacity);
        this.color = d3.scaleOrdinal(d3.schemeCategory10);
        this.node = this.container.append("g").attr("class", "nodes")
            .selectAll("g")
            .data(this.graph.nodes)
            .enter()
            .append("g")
        for (let i = 0; i < this.genesOfInterest.length; i++) {
            this.node.filter(d => d.countPerTrait[i] > 0)
                .append("path")
                .attr("d", d3.arc()
                    .innerRadius(d => d.radiusPerTrait[i] * (this.basePieSize - this.pieBorderSize))
                    .outerRadius(d => d.radiusPerTrait[i] * this.basePieSize)
                    .startAngle(0)
                    .endAngle(Math.PI * 2))
                .attr("fill", this.color(i));
        }
        if (this.isDrawingPies()) {
            this.genesOfInterest.forEach((id, idx) => this.drawPies(idx, id));
        }
        this.node
            .on("click", (event, d) => {
                this.nodeSelector.value = d.id;
                this.edgeSelector.value = "";
                focus(event, d);
                this.mouseoverEnabled = false;
                console.log('clicked', d.id);
                // Re-enable mouseovers after a delay
                setTimeout(() => {
                    this.mouseoverEnabled = true;
                }, 1500);
            })
            .on("mouseover", focus)
        this.node.call(
            d3.drag()
                .on("start", dragstarted)
                .on("drag", dragged)
                .on("end", dragended)
        );
        this.labelNode = this.container.append("g").attr("class", "labelNodes")
            .selectAll("text")
            .data(this.label.nodes)
            .enter()
            .append("text")
            .text(function (d, i) {
                return i % 2 === 0 ? "" : d.node.id;
            })
            .style("fill", "#555")
            .style("font-family", "Arial")
            .style("font-size", this.fontSize)
            .style("pointer-events", "none"); // to prevent mouseover/drag capture
        // Per-type markers, as they don't inherit styles.
        this.svg.append("defs").selectAll("marker")
            .data(["pp", "pp_red", "pp_blue", "pd", "sigma", "met", "metabolic", "srna", "phosphorylation", "PDActivator", "PDRepressor", "PDUnknown", "srnaRepression", "srnaActivation", "srnaUnknown"])
            .enter().append("marker")
            .attr("id", function (d) {
                return d;
            })
            .attr("viewBox", "0 -5 10 10")
            .attr("refX", this.radius * 1.5)
            .attr("orient", "auto")
            .attr("markerUnits", "userSpaceOnUse")
            .attr("markerWidth", this.radius * 2)
            .attr("markerHeight", this.radius * 2)
            .append("path")
            .attr("d", "M0,-5L10,0L0,5");
        // These are the markers from undirected interactions
        this.svg.append("defs").selectAll("marker")
            .data(["ppundir", "metundir"])
            .enter().append("marker")
            .attr("id", function (d) {
                return d;
            })
            .attr("viewBox", "-10 -5 10 10")
            .attr("refX", -this.radius * 1.5)
            .attr("refY", 0)
            .attr("markerUnits", "userSpaceOnUse")
            .attr("markerWidth", this.radius * 2)
            .attr("markerHeight", this.radius * 2)
            .attr("orient", "auto")
            .append("path")
            .attr("d", "M0,-5L-10,0L0,5");
    };
    parsePaths() {
        const pathsByEdge = {};
        const pathsByNode = {};
        for (const pathType in this.paths) {
            for (const path of this.paths[pathType]) {
                const parts = path.split("\t");
                const fromSample = parts[0];
                const toSample = parts[1];
                const score = parseFloat(parts[2]);
                const edgeString = parts[3];
                const {nodes, edges} = this.parseDirectedPath(edgeString);
                const pathInfo = {pathType, fromSample, toSample, score, edgeString, nodes, edges};
                for (const edge of edges) {
                    if (!pathsByEdge[edge]) {
                        pathsByEdge[edge] = [];
                    }
                    pathsByEdge[edge].push(pathInfo);
                }
                for (const node of nodes) {
                    if (!pathsByNode[node]) {
                        pathsByNode[node] = [];
                    }
                    pathsByNode[node].push(pathInfo);
                }
            }
        }
        this.pathsByEdge = pathsByEdge;
        this.pathsByNode = pathsByNode;
    }
    parseDirectedPath(pathStr) {
        const nodes = pathStr.split(/->|<-/);
        const directions = [...pathStr.matchAll(/->|<-/g)].map(d => d[0]);
        const edges = [];
        for (let i = 0; i < directions.length; i++) {
            const from = directions[i] === '->' ? nodes[i] : nodes[i + 1];
            const to = directions[i] === '->' ? nodes[i + 1] : nodes[i];
            edges.push(`${from};${to}`);
        }
        return {
            nodes,
            edges
        };
    }
    getPiePieces(prop) {
        const propIdx = this.genesOfInterest.indexOf(prop);
        return Math.max(...this.graph.nodes.map(n => n.samples[propIdx].length));
    }
    addHighlightButtons() {
        const highlightButtons = document.createElement("div");
        const modeLabel = document.createElement("div");
        modeLabel.textContent = 'Highlight mode on node hover:';
        highlightButtons.appendChild(modeLabel);
        const modeDropdown = document.createElement("select");
        ['paths', 'component', 'neighbors'].forEach(mode => {
            const option = document.createElement("option");
            option.value = mode;
            option.text = mode;
            modeDropdown.appendChild(option);
        });
        modeDropdown.onchange = () => {
            this.highlightMode = modeDropdown.value;
        };
        highlightButtons.appendChild(modeDropdown);
        // 🎯 Edge selector
        const edgeSelectorLabel = document.createElement("div");
        edgeSelectorLabel.textContent = "Select edge:";
        highlightButtons.appendChild(edgeSelectorLabel);
        const edgeSelector = document.createElement("select");
        edgeSelector.id = "edgeSelector";
        edgeSelector.appendChild(new Option("(none)", ""));
        [...this.graph.links].sort((a, b) =>
            `${a.source.id || a.source};${a.target.id || a.target}`.localeCompare(
                `${b.source.id || b.source};${b.target.id || b.target}`
            )
        ).forEach(link => {
            const label = `${link.source} -> ${link.target}`;
            const value = `${link.source};${link.target}`;
            edgeSelector.appendChild(new Option(label, value));
        });
        edgeSelector.onchange = () => {
            const link = this.edgeIndex[edgeSelector.value];
            if (link) {
                this.highlightMode = 'paths';
                edgeFocus(null, link);
            } else {
                clearHighlight();
            }
        };
        highlightButtons.appendChild(edgeSelector);
        this.edgeSelector = edgeSelector;
        // 🎯 Node selector
        const nodeSelectorLabel = document.createElement("div");
        nodeSelectorLabel.textContent = "Select node:";
        highlightButtons.appendChild(nodeSelectorLabel);
        const nodeSelector = document.createElement("select");
        nodeSelector.id = "nodeSelector";
        nodeSelector.appendChild(new Option("(none)", ""));
        [...this.graph.nodes].sort((a, b) => a.id.localeCompare(b.id)).forEach(node => {
            nodeSelector.appendChild(new Option(node.id, node.id));
        });
        nodeSelector.onchange = () => {
            const value = nodeSelector.value;
            const node = this.graph.nodes.find(n => n.id === value);
            if (node) {
                const dummy = {target: {__data__: node}}; // mimic d3 event target
                this.highlightMode = modeDropdown.value; // respect mode
                focus(dummy, node);
            } else {
                clearHighlight();
            }
        };
        highlightButtons.appendChild(nodeSelector);
        this.nodeSelector = nodeSelector;
        document.getElementById('buttons').appendChild(highlightButtons);
    }
    addPieButtons(pieNames) {
        //TODO: check boxes to determine which pies to show
    }
    addFilterButtons(textContent, ids, method) {
        const buttons = document.createElement("div");
        const label = document.createElement("div");
        label.textContent = textContent;
        buttons.appendChild(label);
        // Create a select element for dropdown menu
        const dropdown = document.createElement("select");
        for (const id in ids) {
            // Create an option for each id
            const option = document.createElement("option");
            option.value = id;
            option.text = ids[id];
            dropdown.appendChild(option);
        }
        // Set the onchange event to highlight the selected id
        dropdown.onchange = () => {
            method.apply(this, [dropdown.value]);
        };
        buttons.appendChild(dropdown);
        document.getElementById('buttons').appendChild(buttons);
    }
    addGeneSetButtons() {
        const geneSets = {};
        for (const setID in this.geneSets) {
            geneSets[setID] = setID;
        }
        this.addFilterButtons(
            'Select gene set to highlight:',
            geneSets,
            this.highlightGeneSet,
        );
    }
    highlightGeneSet(setID) {
        clearHighlight();
        const hasGeneSet = d => d.geneSets[setID] ?? false;
        this.node.style("opacity", function (d) {
            return Math.max(0.1, (hasGeneSet(d) ? d.opacity : 0.1));
        });
        this.labelNode.attr("display", function (d, i) {
            if (i % 2 === 0) {
                return 'none';
            }
            //const node = gonetic.graph.nodes[Math.floor(i / 2)]
            return hasGeneSet(d.node) ? "block" : "none";
        });
        this.link.style("opacity", function (d) {
            return hasGeneSet(d.source) && hasGeneSet(d.target)
                ? d.opacity
                : 0.1;
        });
        this.setInfo(`Highlighting gene set ${setID}`);
    }
    addSampleButtons() {
        const samples = {};
        for (const sampleID in this.graph.conditions) {
            samples[sampleID] = this.graph.conditions[sampleID];
        }
        this.addFilterButtons(
            'Select sample to highlight:',
            samples,
            this.highlightSample,
        );
    }
    hasSample(d, sample) {
        for (const path of d.paths) {
            if (path.fromSample === sample || path.toSample === sample) {
                return true;
            }
        }
        return false;
    }
    highlightSample(sampleIdx) {
        clearHighlight();
        const sample = this.conditions[sampleIdx];
        this.link.style("opacity", function (d) {
            if (this.hasSample(d, sample)) {
                return d.opacity;
            }
            return 0.1;
        });
        console.log(nodes)
        this.node.style("opacity", function (d) {
            return Math.max(0.1, (this.hasSample(d, sample) ? d.opacity : 0.1));
        });
        this.labelNode.attr("display", function (d, i) {
            if (i % 2 === 0) {
                return 'none';
            }
            return this.hasSample(d.node, sample) ? "block" : "none";
        });
        this.setInfo(`Highlighting sample ${this.graph.conditions[sampleIdx]}`);
    }
    setInfo(idText, information = '', productInformation = '', geneSetInformation = '') {
        this.infoElement.children[0].textContent = idText;
        this.infoElement.children[1].textContent = information;
        this.infoElement.style.visibility = "visible";
        this.infoElement.style.whiteSpace = "pre-line";
    }
    appendRule = (selector = '*', style = {}) => {
        const sheet = this.styleSheet;
        const len = sheet.cssRules.length;
        sheet.insertRule('*{}', len);
        const rule = sheet.cssRules[len];
        rule.selectorText = selector;
        Object.getOwnPropertyNames(style).forEach(prop => {
            rule.style[prop] = style[prop];
        });
        return rule;
    }
    createStyleSheet() {
        // create style sheet
        document.body.appendChild(document.createElement('style'));
        this.styleSheet = document.styleSheets[document.styleSheets.length - 1];
        // style rules for nodes
        this.appendRule('.node', {
            fill: '#555555',
            'fill-opacity': '0.8',
        });
        // style rules for links
        this.appendRule('.link', {
            fill: 'none',
            'stroke-opacity': '0.7',
            'stroke-width': '' + this.strokeWidth + 'px',
        });
        // style rules for arcs TODO
        this.appendRule('.arc', {
            fill: 'white',
        });
        // link styles
        // TODO: automate this based on presence in the data
        [
            // pp
            {names: ['pp'], color: 'green', opacity: 0.9},
            {names: ['pp_red'], color: 'red', opacity: 0.9},
            {names: ['pp_blue'], color: 'blue', opacity: 0.9},
            // pd
            {names: ['pd', 'PDRepressor'], color: 'red', opacity: 0.9},
            {names: ['PDActivator'], color: 'blue', opacity: 0.9},
            {names: ['PDUnknown'], color: 'LightGray', opacity: 0.9},
            // (de)phosphorylation
            {names: ['phosphorylation'], color: 'purple', opacity: 0.9},
            {names: ['dephosphorylation'], color: 'brown', opacity: 0.9},
            // met
            {names: ['met', 'metabolic'], color: 'orange', opacity: 0.9},
            // srna
            {names: ['srna'], color: 'black', opacity: 0.9},
            // sigma
            {names: ['sigma'], color: 'steelblue', opacity: 0.9},
        ].forEach(linkStyle => {
            const color = linkStyle.color;
            const opacity = linkStyle.opacity;
            linkStyle.names.forEach(name => {
                // link style based on class
                this.appendRule(`.link.${name}`, {
                    stroke: color,
                });
                // marker styles based on id TODO
                [name, `${name}undir`].forEach(id => {
                    this.appendRule(`#${id}`, {
                        fill: color,
                        opacity: opacity,
                    });
                });
            });
        });
    }
    isDrawingPies(prop) {
        // determine if we should draw pies
        for (const id of this.genesOfInterest) {
            if (this.getPiePieces(id) > this.pieLimit) {
                return false;
            }
        }
        return true;
    }
    drawPies(round = 0, prop) {
        if (round !== 0) {
            // draw border between different pie layers
            this.node.append("path")
                .attr("d", d3.arc()
                    .innerRadius(d => d.radius * (this.basePieSize + this.pieSizeIncrease * round - this.pieBorderSize))
                    .outerRadius(d => d.radius * (this.basePieSize + this.pieSizeIncrease * round))
                    .startAngle(0)
                    .endAngle(Math.PI * 2))
                .attr("fill", "white")
        }
        const pieces = this.getPiePieces(prop);
        const arc = (piece) => d3.arc()
            .innerRadius(d => d.radius * (this.basePieSize + this.pieSizeIncrease * round))
            .outerRadius(d => d.radius * (this.basePieSize + this.pieSizeIncrease * (round + 1)))
            .startAngle(piece * Math.PI * 2 / pieces)
            .endAngle((piece + 1) * Math.PI * 2 / pieces)
        for (let i = 0; i < pieces; i++) {
            this.node.append("path")
                .attr("d", arc(i))
                .attr("fill", this.color(i))
                .attr("opacity", n => n.samples[prop][i] ? 1 : 0)
        }
    }
    unfreeze() {
        this.graph.nodes.forEach(d => {
            d.fx = null;
            d.fy = null;
        });
    }
}
window.gonetic = new GoNetic();
const gonetic = window.gonetic;
function centroid(nodes) {
    let x = 0;
    let y = 0;
    let z = 0;
    for (const d of nodes) {
        let k = gonetic.radius ** 2;
        x += d.x * k;
        y += d.y * k;
        z += k;
    }
    return {x: x / z, y: y / z};
}
function forceCluster() {
    const strength = 0.07;
    let nodes;
    function force(alpha) {
        const centroids = d3.rollup(nodes, centroid, d => d.group);
        const l = alpha * strength;
        for (const d of nodes) {
            const {x: cx, y: cy} = centroids.get(d.group);
            d.vx -= (d.x - cx) * l;
            d.vy -= (d.y - cy) * l;
        }
    }
    force.initialize = _ => nodes = _;
    return force;
}
function forceCollide() {
    const alpha = 0.5; // fixed for greater rigidity!
    const padding1 = gonetic.radius; // separation between same-color nodes
    const padding2 = gonetic.radius * 4; // separation between different-color nodes
    let nodes;
    let maxRadius;
    function force() {
        const quadtree = d3.quadtree(nodes, d => d.x, d => d.y);
        for (const d of nodes) {
            const r = gonetic.radius + maxRadius;
            const nx1 = d.x - r, ny1 = d.y - r;
            const nx2 = d.x + r, ny2 = d.y + r;
            quadtree.visit((q, x1, y1, x2, y2) => {
                if (!q.length) do {
                    if (q.data !== d) {
                        const r = 2 * gonetic.radius + (d.group === q.data.group ? padding1 : padding2);
                        let x = d.x - q.data.x, y = d.y - q.data.y, l = Math.hypot(x, y);
                        if (l < r) {
                            l = (l - r) / l * alpha;
                            d.x -= x *= l, d.y -= y *= l;
                            q.data.x += x, q.data.y += y;
                        }
                    }
                } while (q = q.next);
                return x1 > nx2 || x2 < nx1 || y1 > ny2 || y2 < ny1;
            });
        }
    }
    force.initialize = _ => {
        nodes = _;
        maxRadius = 20 + Math.max(padding1, padding2);
    }
    return force;
}
const linkForceStrength = link => {
    if (gonetic.nodeOutDegrees[link.target.id] === 1) {
        return 0.9;
    }
    if (gonetic.nodeOutDegrees[link.source.id] > 5 && gonetic.nodeOutDegrees[link.target.id] < 3) {
        return 0.7;
    } else if (gonetic.nodeOutDegrees[link.source.id] < 4 || gonetic.nodeOutDegrees[link.target.id] < 4) {
        return 0.3;
    } else if (gonetic.nodeOutDegrees[link.source.id] > 10 || gonetic.nodeOutDegrees[link.target.id] > 10) {
        return 0.3;
    }
    return 0.4;
}
const neigh = (a, b) => {
    return a === b || gonetic.adjList[a + "-" + b];
};
const ticked = (event) => {
    gonetic.node.call(updateNode);
    gonetic.link.call(updateLink);
    gonetic.link
        .on("click", (event, d) => {
            gonetic.edgeSelector.value = d.source.id + ";" + d.target.id;
            gonetic.nodeSelector.value = "";
            edgeFocus(event, d);
            gonetic.mouseoverEnabled = false;
            console.log('clicked', d.source.id + ";" + d.target.id);
            // Re-enable mouseovers after a delay
            setTimeout(() => {
                gonetic.mouseoverEnabled = true;
            }, 1500);
        })
        .on("mouseover", edgeFocus)
    gonetic.labelLayout.alphaTarget(0.3).restart();
    gonetic.labelNode.each(function (d, i) {
        if (i % 2 === 0) {
            d.x = d.node.x;
            d.y = d.node.y;
        } else {
            const b = this.getBBox();
            const diffX = d.x - d.node.x;
            const diffY = d.y - d.node.y;
            const dist = Math.sqrt(diffX * diffX + diffY * diffY);
            let shiftX = b.width * (diffX - dist) / (dist * 2);
            shiftX = Math.max(-b.width, Math.min(0, shiftX));
            const shiftY = 16;
            this.setAttribute("transform", "translate(" + shiftX + "," + shiftY + ")");
        }
    });
    gonetic.labelNode.call(updateNode);
}
const fixna = (x) => {
    return isFinite(x) ? x : 0;
}
const updateLink = (link) => {
    link.attr("d", linkArc)
        .attr("marker-end", d => "url(#" + d.type + ")")
        .attr("marker-start", function (d) {
            if (d.direction === "undirected") {
                return "url(#" + d.type + "undir" + ")";
            }
        });
}
// Add different arc bendings for each data type so paths never overlap
const linkArc = d => {
    const dx = d.target.x - d.source.x;
    const dy = d.target.y - d.source.y;
    const distance = Math.sqrt(dx * dx + dy * dy);
    let multiplier = 1;
    switch (d.type) {
        case 'PDRepressor':
        case 'PDActivator':
        case 'PDUnknown':
        case 'pp':
            multiplier = 3;
            break;
        case 'pp_red':
            multiplier = 3;
            break;
        case 'pp_blue':
            multiplier = 3;
            break;
        case 'met':
        case 'metabolic':
            multiplier = 1.75;
            break;
        case 'pd':
            multiplier = 1.25;
            break;
        case 'sigma':
            multiplier = 1;
            break;
        case 'srna':
            multiplier = 0.75;
            break;
        case 'phosphorylation':
            multiplier = 0.85;
            break;
        case 'dephosphorylation':
            multiplier = 0.675;
            break;
    }
    // halve radius multiplier for clearer bends
    multiplier /= 2;
    // determine radius
    const dr = multiplier * distance;
    return "M" + d.source.x + "," + d.source.y + "A" + dr + "," + dr + " 0 0,1 " + d.target.x + "," + d.target.y;
}
let focus = (event, d) => {
    if (!gonetic.mouseoverEnabled) {
        return;
    }
    if (gonetic.highlightMode === 'paths') {
        const paths = gonetic.pathsByNode[d.id] || [];
        highlightPaths(d.id, paths);
        return;
    }
    if (gonetic.nodeSelector) {
        gonetic.nodeSelector.value = d.id;
    }
    const datum = d3.select(event.target).datum();
    let compare = (_) => true;
    let componentConditions = new Set();
    switch (gonetic.highlightMode) {
        case 'neighbors':
            compare = d => neigh(datum.index, d.index);
            break;
        case 'component':
            compare = d => datum.group === d.group;
            break;
    }
    if (gonetic.highlightMode === 'component') {
        gonetic.graph.links.forEach(link => {
            if (link.source.group === datum.group || link.target.group === datum.group) {
                link.paths.forEach(p => componentConditions.add(p.fromSample));
            }
        });
    }
    gonetic.node.style("opacity", function (d) {
        return Math.max(0.1, (compare(d) ? d.opacity : 0.1));
    });
    gonetic.labelNode.attr("display", function (d, i) {
        if (i % 2 === 0) {
            return 'none';
        }
        //const node = gonetic.graph.nodes[Math.floor(i / 2)]
        return compare(d.node) ? "block" : "none";
    });
    gonetic.link.style("opacity", function (d) {
        return compare(d.source) && compare(d.target)
            ? d.opacity
            : 0.1;
    });
    const genesOfInterest = [];
    for (const id in d.samples) {
        const conditions = d.samples[id]
            .map((x, i) => x ? window.graph.conditions[i] : undefined)
            .filter(i => i !== undefined);
        if (conditions.length === 0) {
            genesOfInterest.push(`No ${window.graph.genesOfInterest[id]}.`);
        } else {
            genesOfInterest.push(`${window.graph.genesOfInterest[id]} in: ` + conditions.join(', ') + '.');
        }
    }
    if (d.product !== undefined) {
        genesOfInterest.push('Product: ' + d.product);
    }
    if (gonetic.geneSets !== undefined) {
        const geneSets = [];
        for (const setID in d.geneSets) {
            geneSets.push(setID);
        }
        if (geneSets.length === 0) {
            genesOfInterest.push(`${d.id} is not present in any gene sets.`);
        } else if (geneSets.length === 1) {
            genesOfInterest.push(`${d.id} is present in gene set ${geneSets.join(', ')}.`);
        } else {
            genesOfInterest.push(`${d.id} is present in gene sets ${geneSets.join(', ')}.`);
        }
    } else {
        genesOfInterest.push('No gene sets were provided to GoNetic.');
    }
    if (gonetic.highlightMode === 'component' && componentConditions.size > 0) {
        genesOfInterest.push('Conditions with links in the component: ' + Array.from(componentConditions).join(', ') + '.');
    }
    gonetic.setInfo(
        `Selected node: ${d.id}`,
        genesOfInterest.join('\n'),
    );
}
function highlightPaths(id, paths) {
    clearHighlight();
    const nodesToShow = new Set();
    const linksToShow = new Set();
    const infoBySample = {};
    for (const path of paths) {
        const sample = path.fromSample;
        const type = path.pathType;
        if (!infoBySample[sample]) {
            infoBySample[sample] = {};
        }
        if (!infoBySample[sample][type]) {
            infoBySample[sample][type] = [];
        }
        if (path.toSample !== sample) {
            infoBySample[sample][type].push(`${path.edgeString} (${path.score}) (-> ${path.toSample})`);
        } else {
            infoBySample[sample][type].push(`${path.edgeString} (${path.score})`);
        }
        path.edges.forEach(e => {
            const link = gonetic.edgeIndex[e];
            if (link) linksToShow.add(link);
        });
        path.nodes.forEach(n => nodesToShow.add(n));
    }
    gonetic.node.style("opacity", d => nodesToShow.has(d.id) ? d.opacity : 0.1);
    gonetic.link.style("opacity", d => linksToShow.has(d) ? d.opacity : 0.1);
    gonetic.labelNode.attr("display", function (d, i) {
        if (i % 2 === 0) return 'none';
        return nodesToShow.has(d.node.id) ? 'block' : 'none';
    });
    const infoLines = [];
    for (const sample in infoBySample) {
        infoLines.push(`\n${sample}`);
        for (const type in infoBySample[sample]) {
            for (const str of infoBySample[sample][type]) {
                infoLines.push(`[${type}] ${str}`);
            }
        }
    }
    gonetic.setInfo(`Highlighted  entity ${id} has paths:`, infoLines.join("\n"));
}
function clearHighlight() {
    gonetic.labelNode.attr("display", "block");
    gonetic.node.style("opacity", d => d.opacity);
    gonetic.link.style("opacity", d => d.opacity);
    gonetic.setInfo(gonetic.defaultInfo);
}
function edgeFocus(event, link) {
    if (!gonetic.mouseoverEnabled) {
        return;
    }
    if (gonetic.highlightMode !== 'paths') {
        return;
    }
    highlightPaths(`${link.source.id}->${link.target.id}`, link.paths || []);
}
const updateNode = (node) => {
    node.attr("transform", function (d) {
        return "translate(" + fixna(d.x) + "," + fixna(d.y) + ")";
    });
}
const dragstarted = (event, d) => {
    event.sourceEvent.stopPropagation();
    if (!event.active) {
        gonetic.graphLayout.alphaTarget(0.3).restart();
    }
    d.fx = d.x;
    d.fy = d.y;
}
const dragged = (event, d) => {
    d.fx = event.x;
    d.fy = event.y;
}
const dragended = (event, d) => {
    if (!event.active) {
        gonetic.graphLayout.alphaTarget(0);
    }
}
// presence or absence of mutations or differential expression in a specific gene is encoded as a boolean array
// where each boolean corresponds with a specific sample
// each of these boolean arrays are encoded as a base64 string in the highestScoringSubnetwork.js file
// this file contains functions to convert these base64 strings back to boolean arrays in JavaScript
// base64ToBools converts a base64 string to a boolean array
function base64ToBools(base64, numBools, skip = 0) {
    if (!base64 || base64.length === 0) {
        return new Array(numBools).fill(false);
    }
    return bytesToBools(base64ToBytes(base64, skip), numBools);
}
// base64ToBytes converts a base64 string to a byte array
function base64ToBytes(base64, skip) {
    const binaryString = atob(base64);
    const bytes = new Uint8Array(binaryString.length);
    for (let i = skip; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i);
    }
    return bytes;
}
// bytesToBools converts a byte array to a boolean array
function bytesToBools(bytes, numBools) {
    const bools = new Array(numBools);
    for (let i = 0; i < numBools; i++) {
        bools[i] = (bytes[Math.floor(i / 8)] & (1 << (7 - (i % 8)))) !== 0;
    }
    return bools;
}
// wait for graph and d3 to load
init = () => {
    if (!window.graph || !window.d3) {
        setTimeout(init, 100);
    }
    window.gonetic.init(
        window.graph,
        window.paths,
        window.geneSets,
        7.5,
    );
}
setTimeout(init, 100);