hljs.initHighlightingOnLoad();

$(function() {
    if ($('.here').length > 0) {
        var percentagedown = ($('.here').position().top / $(window).height()) * 100;
        if (percentagedown > 50) {
            var value = $('.here').position().top - ($(window).height() / 2) + ($('nav ul li:first').height() / 2);
            $(".nano").nanoScroller({
                scrollTop: value
            });
        } else {
            $(".nano").nanoScroller();
        }
    }
    $('time.timeago').timeago();
    $('button.svgsave').on('click', function() {
        var html = d3.select("svg")
            .attr("version", 1.1)
            .attr("xmlns", "http://www.w3.org/2000/svg")
            .node().parentNode.innerHTML;

        //console.log(html);
        var imgsrc = 'data:image/svg+xml;base64,' + btoa(html);
        var img = '<img src="' + imgsrc + '">';
        d3.select("#svgdataurl").html(img);


        var canvas = document.querySelector("canvas"),
            context = canvas.getContext("2d");

        var image = new Image();
        image.src = imgsrc;
        image.onload = function() {
            context.drawImage(image, 0, 0);

            var canvasdata = canvas.toDataURL("image/png");

            var pngimg = '<img src="' + canvasdata + '">';
            d3.select("#pngdataurl").html(pngimg);

            var a = document.createElement("a");
            a.download = "sample.png";
            a.href = canvasdata;
            a.click();
        };
    });
});

var growse = {
    map: {
        width: 816,
        height: 480,
        svg: null,
        path: null,
        zoom: null,
        move: function() {
            var g = growse.map.svg.select('.mapgroup');
            var t = d3.event.translate;
            var s = d3.event.scale;
            growse.map.zoom.translate(t);
            g.attr("transform", "translate(" + t + ")scale(" + s + ")");
            d3.selectAll(".route path").style("stroke-width", (1.5 / s) + "px");

        }
    },
    projection: null,
    mapFeature: null,

    drawMap: function(elemId, mapfile) {
        $('select[name=year]').on('change', function() {
            console.log($(this).val());
            growse.drawRoute($(this).val());
        });
        //Initial dimensions
        //Set up the map projection
        growse.projection = d3.geo.equirectangular()
            .center([0, 0])
            .scale(140)
            .translate([growse.map.width / 2, growse.map.height / 2])
            .precision(0.1);

        //Turn the projection into a path
        growse.map.path = d3.geo.path()
            .projection(growse.projection);

        growse.map.zoom = d3.behavior.zoom()
            .scaleExtent([1, 500])
            .on("zoom", growse.map.move);
        growse.map.svg = d3.select(elemId).append("svg")
            .attr("width", growse.map.width)
            .attr("height", growse.map.height)
            .call(growse.map.zoom);

        d3.json(mapfile, function(error, world) {
            var g = growse.map.svg.append('g').attr('class', 'mapgroup');
            g.append("path")
                .datum(topojson.feature(world, world.objects.land))
                .attr("class", "land")
                .attr("d", growse.map.path)
                .style("fill", "#57d");
            growse.drawRoute($("select[name=year]").val());

        });

        d3.select(self.frameElement).style("height", growse.map.height + "px");
        d3.select(self.frameElement).style("width", growse.map.width + "px");

    },
    drawRoute: function(year) {
        var g = growse.map.svg.select(".mapgroup");
        d3.selectAll(".route").remove();
        d3.json("/where/linestring/" + year + "/", function(error, mypath) {
            //Add the path
            g.append("g")
                .attr("class", "route")
                .selectAll("path")
                .data(mypath.features)
                .enter()
                .append("path")
                .attr("d", growse.map.path)
                .style("fill-opacity", "0.0")
                .style("fill", "#000")
                .style("stroke", "#f90")
                .style("stroke-width", "1px");

            /*
            var targetPath = d3.selectAll('.route')[0][0];
            var pathNode = d3.select(targetPath).selectAll('path').node();
            var pathLength = pathNode.getTotalLength();
            d3.select('.route')
                .style('stroke-dasharray', pathLength)
                .style('stroke-dashoffset', 0)
                .style('-webkit-animation', "flarble 60s linear forwards");*/
        });
    },
    drawLineChart: function(elemId, data, width, height, xAxisTitle, yAxisTitle) {
        var xpadding = 45;
        var ypadding = 15;
        if (!data) {
            return;
        }
        //var x = d3.scale.linear().range([0, width]);
        var x = d3.time.scale().range([xpadding, width - (xpadding * 2)]);
        var y = d3.scale.linear().range([height - (ypadding * 2), ypadding]);
        var line = d3.svg.line()
            .interpolate("basis")
            .x(function(d) {
                return x(d.date);
            })
            .y(function(d) {
                return y(d.val);
            });
        data.forEach(function(d) {
            d.date = new Date(d[0] * 1000);
            d.val = d[1];
        });
        x.domain(d3.extent(data, function(d) {
            return d.date;
        }));
        y.domain(d3.extent(data, function(d) {
            return d.val;
        }));
        var e = d3.select(elemId);
        var svg = e.append('svg');
        svg.attr('width', width)
            .attr('height', height)
            .append('path')
            .datum(data)
            .attr('class', 'sparkline')
            .attr('d', line);
        svg.append("g")
            .attr("class", "axis")
            .attr("transform", "translate(0," + (height - ypadding - ypadding) + ")")
            .call(d3.svg.axis()
                .scale(x)
                .orient("bottom"));
        svg.append("g")
            .attr("class", "axis")
            .attr("transform", "translate(" + xpadding + ",0)")
            .call(d3.svg.axis()
                .scale(y)
                .orient("left")
                .ticks(5));
        if (yAxisTitle) {
            svg.append("text")
                .attr("transform", "translate(" + xpadding / 2 + "," + height / 2 + ") rotate(270) ")
                .attr('class', 'axislabel')
                .text(yAxisTitle);
        }

    },
    drawColumnChart: function(elemId, data, width, height, xAxisTitle, yAxisTitle) {
        var xpadding = 45;
        var ypadding = 15;
        if (!data) {
            return;
        }
        var x = d3.scale.ordinal().rangeRoundBands([xpadding, width - (xpadding * 2)]);
        var y = d3.scale.linear().range([height - (ypadding * 2), ypadding]);
        data.forEach(function(d) {
            d.date = d[0];
            d.val = d[1];
        });

        x.domain(data.map(function(d) {
            return d.date;
        }));
        y.domain(d3.extent(data, function(d) {
            return d.val;
        }));
        var e = d3.select(elemId);
        var svg = e.append('svg');
        svg.attr('width', width)
            .attr('height', height);
        svg.append("g")
            .selectAll('rect')
            .data(data)
            .enter()
            .append('rect')
            .attr('class', 'chartbar')
            .attr('fill', '#57d')
            .attr("width", function() {
                return 0.8 * x.rangeBand();
            })
            .attr("height", function(d) {
                return (height - (2 * ypadding) - y(d.val)) + "px";
            })
            .attr('x', function(d) {
                return (0.1 * x.rangeBand() + x(d.date)) + "px";

            }).attr('y', function(d) {
                return y(d.val);
            });

        svg.append("g")
            .attr("class", "axis")
            .attr("transform", "translate(0," + (height - ypadding - ypadding) + ")")
            .call(d3.svg.axis()
                .scale(x)
                .orient("bottom"));
        svg.append("g")
            .attr("class", "axis")
            .attr("transform", "translate(" + xpadding + ",0)")
            .call(d3.svg.axis()
                .scale(y)
                .orient("left")
                .ticks(5));
        if (yAxisTitle) {
            svg.append("text")
                .attr("transform", "translate(" + xpadding / 2 + "," + height / 2 + ") rotate(270) ")
                .attr('class', 'axislabel')
                .text(yAxisTitle);
        }
        if (xAxisTitle) {
            svg.append("text")
                .attr("transform", "translate(" + width / 2 + "," + height + ") ")
                .attr('class', 'axislabel')
                .text(xAxisTitle);
        }
    }
};