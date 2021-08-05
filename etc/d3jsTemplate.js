class GoNetic {
    constructor() {
    };

    init(graph, radius) {
        // keep a copy of arguments
        this.graph = graph;
        this.radius = radius;
        this.highlightMode = 'neighbors';

        // create style sheet
        document.body.appendChild(document.createElement('style'));
        this.styleSheet = document.styleSheets[document.styleSheets.length - 1];
        const appendRule = (selector = '*', style = {}) => {
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

        // style rules for nodes
        appendRule('.node', {
            fill: '#555555',
            'fill-opacity': '0.8',
        });

        // style rules for links
        appendRule('.link', {
            fill: 'none',
            'stroke-opacity': '0.7',
            'stroke-width': '6px',
        });

        // style rules for arcs TODO
        appendRule('.arc', {
            fill: 'white',
        });

        // link styles
        // TODO: automate this based on presence in the data
        [
            // pp
            {names: ['pp'], color: 'green', opacity: 0.9},
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
                appendRule(`.link.${name}`, {
                    stroke: color,
                });
                // marker styles based on id TODO
                [name, `${name}undir`].forEach(id => {
                    appendRule(`#${id}`, {
                        fill: color,
                        opacity: opacity,
                    });
                });
            });
        });

        // determine svg size based on number of nodes
        this.nodeCount = this.graph.nodes.length;
        this.width = 300 * Math.sqrt(this.nodeCount);
        this.height = 200 * Math.sqrt(this.nodeCount);

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
            .force("link", d3.forceLink(label.links).distance(0).strength(2));


        // gather nodes and links from graph data
        this.nodeInDegrees = {};
        this.nodeOutDegrees = {};
        this.neighbors = {};
        this.graph.links.forEach(link => {
            link.weight = -link.max_cost;
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

        this.svg = d3.select("#network")
            .append("div")
            // Container class to make it responsive.
            .classed("svg-container", true)
            .append("svg")
            // Responsive SVG needs these 2 attributes and no width and height attr.
            .attr("preserveAspectRatio", "xMinYMin meet")
            .attr("viewBox", `0 0 ${this.width} ${this.height}`)
            // Class to make it responsive.
            .classed("svg-content-responsive", true);
        appendRule(
            '.svg-container',
            {
                display: 'inline-block',
                position: 'relative',
                width: '100%',
                'padding-bottom': '100%', /* aspect ratio */
                'vertical-align': 'top',
                overflow: 'hidden',
            }
        );
        appendRule('.svg-content-responsive', {
            display: 'inline-block',
            position: 'absolute',
            top: '10px',
            left: '0',
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
            .attr("stroke-width", d => 5 + ((d.max_cost + 1) * 12) / 8)
            .attr("opacity", d => (1 + d.max_cost) / 8)

        this.color = d3.scaleOrdinal(d3.schemeCategory10);

        this.node = this.container.append("g").attr("class", "nodes")
            .selectAll("g")
            .data(this.graph.nodes)
            .enter()
            .append("circle")
            .attr("r", this.radius)
            .attr("fill", colorNode);

        this.node
            .on("mouseover", focus)
            .on("mouseout", unfocus);

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
                return i % 2 == 0 ? "" : d.node.id;
            })
            .style("fill", "#555")
            .style("font-family", "Arial")
            .style("font-size", 24)
            .style("pointer-events", "none"); // to prevent mouseover/drag capture

        // add border
        const border = Math.ceil(1 + Math.sqrt(this.nodeCount));
        const bordercolor = 'black';
        this.svg = this.svg.attr("border", border);
        this.svg.append("rect")
            .attr("x", 0)
            .attr("y", 0)
            .attr("height", this.height)
            .attr("width", this.width)
            .style("stroke", bordercolor)
            .style("fill", "none")
            .style("stroke-width", border);

        // Per-type markers, as they don't inherit styles.
        this.svg.append("defs").selectAll("marker")
            .data(["pp", "pd", "sigma", "met", "metabolic", "srna", "phosphorylation", "PDActivator", "PDRepressor", "PDUnknown", "srnaRepression", "srnaActivation", "srnaUnknown"])
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
        /*
        this.pies = this.svg.append("svg:g").selectAll("svg")
            .data(this.force.nodes())
            .enter()
            .append("g")
            .call(this.node_drag)
            .attr("class", "arc");
        */
    };

    unfreeze() {
        this.graph.nodes.forEach(d => {
            d.fx = null;
            d.fy = null;
        });
    }

    /*
    tick() {
        const forceBoundaryX = (x) => this.forceBoundary(x, this.radius, this.w);
        const forceBoundaryY = (y) => this.forceBoundary(y, this.radius, this.h);
        this.path.attr("d", this.linkArc);
        this.circles.attr("transform", function (d) {
            d.x = forceBoundaryX(d.x);
            d.y = forceBoundaryY(d.y);
            return "translate(" + d.x + "," + d.y + ")";
        });
        this.pies.attr("transform", function (d) {
            return "translate(" + d.x + "," + d.y + ")";
        });
        this.text.attr("transform", function (d) {
            return "translate(" + d.x + "," + d.y + ")";
        });
    }


    // rgb interpolators
    this.interpolatorPositive = d3.interpolateRgb("white", "red");
    this.interpolatorNegative = d3.interpolateRgb("white", "green");
    addMutatorArc(sampleIndex, color, sampleCount) {
        const radians = 2 * Math.PI / sampleCount;
        const arc = d3.svg.arc()
            .outerRadius(this.radius)
            .innerRadius(this.radius * 2 / 3)
            .startAngle(sampleIndex * radians) //converting from degs to radians
            .endAngle((sampleIndex + 1) * radians);
        pies.append("path")
            .attr("d", arc)
            .style("fill", d => d.mutant[sampleIndex] ? color : "gray");
    }

    renderLFCColor(value) {
        const result = Math.min(Math.max(0, (Math.abs(value)) / 2.5), 1);
        return value < 0
            ? this.interpolatorNegative(result)
            : this.interpolatorPositive(result);
    }

    addDifferentialExpressionArc(sampleIndex, sampleCount) {
        const radians = 2 * Math.PI / sampleCount;
        const arc = d3.svg.arc()
            .outerRadius(this.radius * 2 / 3)
            .innerRadius(0)
            .startAngle(sampleIndex * radians) //converting from degs to radians
            .endAngle((sampleIndex + 1) * radians);
        pies.append("path")
            .attr("d", arc)
            .style("fill", d => {
                const lfc = d.lfcs[sampleIndex];
                return lfc ? this.renderLFCColor(lfc) : "gray";
            });
    }
    */
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

const colorNode = (d) => gonetic.color(d.group);

const ticked = (event) => {
    gonetic.node.call(updateNode);
    gonetic.link.call(updateLink);

    gonetic.labelLayout.alphaTarget(0.3).restart();
    gonetic.labelNode.each(function (d, i) {
        if (i % 2 == 0) {
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

const focus = (event, d) => {
    const datum = d3.select(event.target).datum();
    let compare = (_) => true;
    switch (gonetic.highlightMode) {
        case 'neighbors':
            compare = d => neigh(datum.index, d.index);
            break;
        case 'component':
            compare = d => datum.group === d.group;
            break;
    }
    gonetic.node.style("opacity", function (d) {
        return compare(d) ? 1 : 0.1;
    });
    gonetic.labelNode.attr("display", function (d, i) {
        if (i % 2 === 0) {
            return 'none';
        }
        //const node = gonetic.graph.nodes[Math.floor(i / 2)]
        return compare(d.node) ? "block" : "none";
    });
    gonetic.link.style("opacity", function (d) {
        const opacity = (1 + d.max_cost) / 8
        return compare(d.source) && compare(d.target)
            ? opacity
            : 0.1;
    });
}

const unfocus = (event) => {
    gonetic.labelNode.attr("display", "block");
    gonetic.node.style("opacity", 1);
    gonetic.link.style("opacity", d => (1 + d.max_cost) / 8);
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

init = () => {
    if (!window.graph || !window.d3) {
        setTimeout(init, 100);
    }
    window.gonetic.init(window.graph, 7.5);
}

setTimeout(init, 100);