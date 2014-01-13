// The graph used for the visualization. added to own file
// to separate code for visualizing and keeping track of graph


var nodes = []; 
var edges = []; 
var graph; 

var color = d3.scale.linear()
    .domain([0, 100])
    .range(["yellow", "purple"]);

function Graph(cy){
    this.addNode = function(n){
        
        if(typeof findNode(n.id) != 'undefined') {
            console.log("attempted to add node (",n.id,") which exists.."); 
            return
        }
        

        console.log("Adding node:",n); 
        n.graphics.name = n.graphics.name.split(" ")
        if(n.name === "\"bg\"") {
            console.log("MY LORD!!!!!!!!!!", n.name)
            n.graphics.bgimage = n.graphics.name[0];
            n.graphics.bgcolor = "#fff"
            n.graphics.name = ""
        }
        else{
            n.graphics.bgimage = "";
        }

//        n.background-image = "http://www.genome.jp/kegg/pathway/hsa/hsa04915.png"
        
        // Fetch coloring if node is a gene
        if(n.graphics.shape == "rectangle"){
            // gene name but strip away any colon
            gene = n.name
            id = JSON.parse(n.name)
            ids = JSON.parse(n.name).split(" ")
            
            console.log(gene, ids, ids[0])
            // If more than gene is present in the node. Fetch all of them
            // and do some clever stuff 
            if (ids.length > 0) {
                n.graphics.bgcolor = color(AvgDiff(ids[0]))
            }
            else {

                n.graphics.bgcolor = color(AvgDiff(ids[0]))

            }
        }


        var no = {
            group: 'nodes',
            data: { 
                id: ''+ n.id,
                name: JSON.parse(n.name),
                graphics: n.graphics, 
                //background-image: "test.png",
            },
            position: {
                x: n.graphics.x,
                y: n.graphics.y
            },
            grabbable:false,
        };

        nodes.push(no);
        cy.add(no); 
    };

    this.addEdge = function(e){
        var s = findNode(e.source); 
        var t = findNode(e.target); 
        
        if(typeof s == 'undefined' || typeof t == 'undefined'){
            console.log("Attempted to add a faulty edge"); 
            return
        }

        var ed = {
            group: "edges",
            data: {
                source: ''+e.source,
                target: ''+e.target,
            },
        }; 

        edges.push(ed); 
        //cy.add(ed); 

    }

    var findNode = function (id) {
        for (var i in nodes) {
            if (nodes[i]["data"]["id"] === ''+id) {
                return nodes[i];
            }
        };
    };

    var update = function() {
        //cy.layout(); 
    }
}


