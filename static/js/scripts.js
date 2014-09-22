$(function () {
    if ($('.here').length > 0) {
        var percentagedown = ($('.here').position().top / $(window).height()) * 100;
        if (percentagedown > 50) {
            var value = $('.here').position().top - ($(window).height() / 2) + ($('nav ul li:first').height() / 2);
            $(".nano").nanoScroller({scrollTop: value});
        } else {
            $(".nano").nanoScroller();
        }
    }
});

var growse = {
    drawLineChart: function (elemId, data, width, height, xAxisTitle, yAxisTitle) {
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
            .x(function (d) {
                return x(d.date);
            })
            .y(function (d) {
                return y(d.val);
            });
        data.forEach(function (d) {
            d.date = new Date(d[0] * 1000);
            d.val = d[1];
        });
        x.domain(d3.extent(data, function (d) {
            return d.date;
        }));
        y.domain(d3.extent(data, function (d) {
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
                .orient("bottom"))
        svg.append("g")
            .attr("class", "axis")
            .attr("transform", "translate(" + xpadding + ",0)")
            .call(d3.svg.axis()
                .scale(y)
                .orient("left")
                .ticks(5))
        if (yAxisTitle) {
            svg.append("text")
                .attr("transform", "translate(" + xpadding / 2 + "," + height / 2 + ") rotate(270) ")
                .attr('class', 'axislabel')
                .text(yAxisTitle)
        }

    },
    drawMap: function (elemId) {
        //Initial dimensions
        var width = 816,
            height = 480;
        //Set up the map projection
        var projection = d3.geo.equirectangular()
            .scale(140)
            .translate([width / 2, height / 2])
            .precision(.1);

        //Turn the projection into a path
        var path = d3.geo.path()
            .projection(projection);

        var svg = d3.select("#map").append("svg")
            .attr("width", width)
            .attr("height", height);


        d3.json("/static/js/world-50m.json", function (error, world) {
            svg.append("path")
                .datum(topojson.feature(world, world.objects.land))
                .attr("class", "land")
                .attr("d", path)
                .style("fill", "#57d");
            d3.json("/where/linestring/", function (error, mypath) {
                //Add the path
                svg.append("g")
                    .attr("class", "route")
                    .selectAll("path")
                    .data(mypath.features)
                    .enter()
                    .append("path")
                    .attr("d", path)
                    .style("fill-opacity", "0.0")
                    .style("fill", "#000")
                    .style("stroke", "#f90")
                    .style("stroke-width", "1")

                ;

                var targetPath = d3.selectAll('.route')[0][0];
                var pathNode = d3.select(targetPath).selectAll('path').node();
                var pathLength = pathNode.getTotalLength();
                console.log(pathLength);
                /*

                 //Add the circle
                 var circle =
                 svg.append("circle").attr({
                 r: 3,
                 fill: '#f33',
                 transform: function () {
                 var p = pathNode.getPointAtLength(0)
                 return "translate(" + [p.x, p.y] + ")";
                 }
                 });

                 circle.transition()
                 .duration(duration)
                 .ease("bounce")
                 .attrTween("transform", function (d, i) {
                 return function (t) {
                 var p = pathNode.getPointAtLength(pathLength * t);
                 return "translate(" + [p.x, p.y] + ")";
                 }
                 });
                 */
            });
        });

        d3.select(self.frameElement).style("height", height + "px");
        d3.select(self.frameElement).style("width", width + "px");

    },
    drawColumnChart: function (elemId, data, width, height, xAxisTitle, yAxisTitle) {
        var xpadding = 45;
        var ypadding = 15;
        if (!data) {
            return;
        }
        var x = d3.scale.ordinal().rangeRoundBands([xpadding, width - (xpadding * 2)]);
        var y = d3.scale.linear().range([height - (ypadding * 2), ypadding]);
        data.forEach(function (d) {
            d.date = d[0];
            d.val = d[1];
        });

        x.domain(data.map(function (d) {
            return d.date;
        }));
        y.domain(d3.extent(data, function (d) {
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
            .attr("width", function () {
                return 0.8 * x.rangeBand();
            })
            .attr("height", function (d) {
                return (height - (2 * ypadding) - y(d.val)) + "px";
            })
            .attr('x', function (d) {
                return (0.1 * x.rangeBand() + x(d.date)) + "px";

            }).attr('y', function (d) {
                return y(d.val)
            });

        svg.append("g")
            .attr("class", "axis")
            .attr("transform", "translate(0," + (height - ypadding - ypadding) + ")")
            .call(d3.svg.axis()
                .scale(x)
                .orient("bottom"))
        svg.append("g")
            .attr("class", "axis")
            .attr("transform", "translate(" + xpadding + ",0)")
            .call(d3.svg.axis()
                .scale(y)
                .orient("left")
                .ticks(5))
        if (yAxisTitle) {
            svg.append("text")
                .attr("transform", "translate(" + xpadding / 2 + "," + height / 2 + ") rotate(270) ")
                .attr('class', 'axislabel')
                .text(yAxisTitle)
        }
        if (xAxisTitle) {
            svg.append("text")
                .attr("transform", "translate(" + width / 2 + "," + height + ") ")
                .attr('class', 'axislabel')
                .text(xAxisTitle)
        }

    }
}