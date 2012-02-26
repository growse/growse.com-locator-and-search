//Globals
var growse = {};

growse = function() {
	return {
		loadFrontPage: function() {
			growse.getLocation();
			growse.spamwatchGraph();
			growse.loadRandomPhoto();
			growse.getTweets();
		},
		getTweets: function() {
			$.getJSON("http://res.growse.com/nocache/twitter.js",function(data) {
				var max = 4;
				var counter=0;
				$('#twitter_div').empty();
				var list = $('<ul></ul>');
				$.each(data,function(i,tweet) {
					if (tweet.in_reply_to_status_id == null && counter<max) {
						
						list.append('<li>'+growse.replaceURLWithHTMLLinks(tweet.text)+'<br /><span><a href="https://twitter.com/growse/status/'+tweet.id_str+'">'+prettyDate(tweet.created_at)+'</a></span></li>');
						counter+=1;
					}
				});
				$('#twitter_div').append(list);
			});
		},
		replaceURLWithHTMLLinks: function(text) {
			var exp = /(\b(https?|ftp|file):\/\/[-A-Z0-9+&@#\/%?=~_|!:,.;]*[-A-Z0-9+&@#\/%=~_|])/ig;
			return text.replace(exp,"<a href='$1'>$1</a>");
		},
		getLocation: function() {
			$.getJSON("http://res.growse.com/nocache/latitude.js",function(data) {
					var coords = data.features[0].geometry.coordinates[1]+','+data.features[0].geometry.coordinates[0];
					var url = 'http://maps.googleapis.com/maps/api/staticmap?markers=color:red|'+coords+'&zoom=13&size=285x200&sensor=false';
				$('#twitterlocation_div p').html("<a href=\"http://maps.google.com?q="+coords+"\"><img src="+url+" /></a>");
			});
		},
		getLinks: function() {
			$.getJSON('http://res.growse.com/nocache/links.js', function(data) {
				var linklist={};
				$(data).each(function() {
					var tag = this.t[0];
					var link = [];
					link.description = this.d;
					link.url = this.u;
					if (linklist[tag]==undefined) {
						linklist[tag]=[];
					}
					linklist[tag].push(link);
				});
				$('#links img').remove();
				$.each(linklist,function(key,value) {
					$('#links').append(new jQuery('<h2>'+key+'</h2>'));
					var list = new jQuery('<ul></ul>');
					for (var x=0; x< value.length;x++) {
						list.append(new jQuery('<li><a title=\"'+value[x].description+'\" href=\"'+value[x].url+'\">'+value[x].description+'</a></li>'));
					}
					$('#links').append(list);
				});
			});
		},
		loadRandomPhoto: function() {
			$.getJSON("http://api.flickr.com/services/rest/?method=flickr.photosets.getPhotos&api_key=4115e4c68bbb89a0193cf80bc8554ab2&photoset_id=72157624918832108&privacy_filter=1&extras=url_s%2C+url_m%2C+url_o&format=json&jsoncallback=?", function(data){
				if (data.photoset==undefined) {
					document.write(data.message);
				} else {
					var index = Math.ceil(Math.random() * data.photoset.photo.length);
					var photo = data.photoset.photo[index-1];
					$('#randomphoto').html("<a href=\"/photos#"+index+"\"><img src=\""+photo.url_s+"\" /></a>");
				}
			});
		},
		loadPhotoGallery: function() {
			$.getJSON("http://api.flickr.com/services/rest/?method=flickr.photosets.getPhotos&api_key=4115e4c68bbb89a0193cf80bc8554ab2&photoset_id=72157624918832108&privacy_filter=1&extras=url_sq%2C+url_m%2C+url_o&format=json&jsoncallback=?", function(data){
				if (data.photoset==undefined) {
					document.write(data.message);
				} else {
					$.each(data.photoset.photo, function(i,item){
						var newthumb = $("ul.thumbs").children("li:first").clone();
					
						var thumbimg = item.url_sq;//baseimg.replace("_m.jpg", "_s.jpg");
						$(newthumb).find("img").attr("src", thumbimg);
					
						var disimg = item.url_m;//baseimg.replace("_m.jpg", ".jpg");
						$(newthumb).find(".thumb").attr("href", disimg);
					
						var lgeimg = item.url_o;//baseimg.replace("_m.jpg", "_b.jpg");
						$(newthumb).find(".download").children("a").attr("href", lgeimg);
					
						var title = 'Title';
						var description = 'Desc';
							
						var desc = $("<div />").append(description);
						if ($(desc).children().size() == 3) {
							description = $(desc).children("p:last").html();
						} else {
							description = "";
						}
					
						$(newthumb).find(".image-title").empty().html(title);
						$(newthumb).find(".image-desc").empty().html(description);
						$(newthumb).find(".image-auth").empty().html(item.author);
						$("ul.thumbs").append(newthumb);
					});	
				
					$("ul.thumbs").children("li:first").remove();
					
					// Initially set opacity on thumbs and add
					// additional styling for hover effect on thumbs
					var onMouseOutOpacity = 0.67;
					$('#thumbs ul.thumbs li').opacityrollover({
						mouseOutOpacity: onMouseOutOpacity,
						mouseOverOpacity:1.0,
						fadeSpeed:		 'fast',
						exemptionSelector: '.selected'
					});
					
					// Initialize Advanced Galleriffic Gallery
					var gallery = $('#thumbs').galleriffic({
						delay:					 2500,
						numThumbs:				 15,
						preloadAhead:			10,
						enableTopPager:			true,
						enableBottomPager:		 false,
						maxPagesToShow:			7,
						imageContainerSel:		 '#slideshow',
						controlsContainerSel:	'#controls',
						captionContainerSel:	 '#caption',
						loadingContainerSel:	 '#loading',
						renderSSControls:		false,
						renderNavControls:		 true,
						playLinkText:			'Play Slideshow',
						pauseLinkText:			 'Pause Slideshow',
						prevLinkText:			'&lsaquo; Previous Photo',
						nextLinkText:			'Next Photo &rsaquo;',
						nextPageLinkText:		'Next &rsaquo;',
						prevPageLinkText:		'&lsaquo; Prev',
						enableHistory:			 false,
						autoStart:				 false,
						syncTransitions:		 true,
						defaultTransitionDuration: 900,
						onSlideChange:			 function(prevIndex, nextIndex) {
							// 'this' refers to the gallery, which is an extension of $('#thumbs')
							this.find('ul.thumbs').children()
								.eq(prevIndex).fadeTo('fast', onMouseOutOpacity).end()
								.eq(nextIndex).fadeTo('fast', 1.0);
						},
						onPageTransitionOut:	 function(callback) {
							this.fadeTo('fast', 0.0, callback);
						},
						onPageTransitionIn:		function() {
							this.fadeTo('fast', 1.0);
						}
					});
										
				}
				
				// PageLoad function
				// This function is called when:
				// 1. after calling $.historyInit();
				// 2. after calling $.historyLoad();
				// 3. after pushing "Go Back" button of a browser
				function pageload(hash) {
					//alert("pageload: " + hash);	// hash doesn't contain the first # character.
					if(hash) {
						$.galleriffic.gotoImage(hash);
					} else {
						gallery.gotoIndex(0);
					}
				}

				// Initialize history plugin.
				// The callback is called at once by present location.hash.
				$.historyInit(pageload,'index.html');

				// set onlick event for buttons using the jQuery 1.3 live method
				$("a[rel='history']").live('click', function(e) {
					if (e.button != 0) return true;

					var hash = this.href;
					hash = hash.replace(/^.*#/, '');

					// moves to a new page.
					// pageload is called at once.
					// hash don't contain "#", "?"
					$.historyLoad(hash);
					return false;
				});


				// We only want these styles applied when javascript is enabled
				$('div.navigation').css({'width' : '300px', 'float' : 'left'});
				$('div.content').css('display', 'block');
			});
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

/*
	* JavaScript Pretty Date
	* Copyright (c) 2008 John Resig (jquery.com)
	* Licensed under the MIT license.
	*/

// Takes an ISO time and returns a string representing how
// long ago the date represents.
function prettyDate(time){
	var date = new Date((time || "").replace(/-/g,"/").replace(/[TZ]/g," ")),
		diff = (((new Date()).getTime() - date.getTime()) / 1000),
		day_diff = Math.floor(diff / 86400);
			
	if ( isNaN(day_diff) || day_diff < 0 || day_diff >= 31 )
		return;
			
	return day_diff == 0 && (
			diff < 60 && "just now" ||
			diff < 120 && "1 minute ago" ||
			diff < 3600 && Math.floor( diff / 60 ) + " minutes ago" ||
			diff < 7200 && "1 hour ago" ||
			diff < 86400 && Math.floor( diff / 3600 ) + " hours ago") ||
		day_diff == 1 && "Yesterday" ||
		day_diff < 7 && day_diff + " days ago" ||
		day_diff < 31 && Math.ceil( day_diff / 7 ) + " weeks ago";
}

// If jQuery is included in the page, adds a jQuery plugin to handle it as well
if ( typeof jQuery != "undefined" )
	jQuery.fn.prettyDate = function(){
		return this.each(function(){
			var date = prettyDate(this.title);
			if ( date )
				jQuery(this).text( date );
		});
	};
