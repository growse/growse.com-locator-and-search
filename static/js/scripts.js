//Globals
var growse = {};

growse = function() {
	return {
		getLocation: function() {
			$.getJSON("http://res.growse.com/nocache/latitude.js",function(data) {
					var coords = data.features[0].geometry.coordinates[1]+','+data.features[0].geometry.coordinates[0];
					var url = 'http://maps.googleapis.com/maps/api/staticmap?markers=color:red|'+coords+'&zoom=13&size=285x200&sensor=false';
				$('#twitterlocation_div p').html("<a href=\"http://maps.google.com?q="+coords+"\"><img src="+url+" /></a>");
			});
		},
		clearSearch: function() {
			if ($('#searchbox').val()=='Search') { $('#searchbox').val('');}
		},
		spamwatchGraph: function() {
			$.getJSON("http://res.growse.com/spamwatch/data.work.txt", function (data) {
				var chart = new Highcharts.Chart({
					chart: {
						renderTo: 'spamwatch',
						defaultSeriesType: 'line',
						backgroundColor: '#222222',
						height: 300
					},
					title: {
						text: 'Spams received',
						style: {
							color: '#aaaaaa'
						}
					},
					xAxis: {
						type: 'datetime',
						title: {text: null}
					},
					yAxis: {
						title: {text: null},
						min: 0
					},
					plotOptions: {
						line: {
							marker: {
								enabled: false
							}
					      }
					},
					legend: {enabled: false},
					series: [{
						name: 'spam',
						data:data

					}]
				});
			});
		}
	};
}();
