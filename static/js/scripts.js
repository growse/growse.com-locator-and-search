$(function () {
    var percentagedown = ($('.here').position().top / $(window).height()) * 100;
    if (percentagedown > 50) {
        $('nav').scrollTop($('.here').position().top - ($(window).height() / 2) + ($('nav ul li:first').height() / 2));
    }
    $('nav').bind('mousewheel DOMMouseScroll', function (e) {
        var delta = e.wheelDelta || e.originalEvent.wheelDelta || -e.detail || -e.originalEvent.detail;
        this.scrollTop += ( delta < 0 ? 1 : -1 ) * 40;
        //This gets around a bug in chrome beta where the nav doesn't rerender after scrollTop()
        this.style.display = 'none';
        this.offsetHeight;
        this.style.display = 'block';
        e.preventDefault();
    });
});

var growse = function () {
    return {
        loadingNav: false,
        loadNav: function () {
            if (!growse.loadingNav) {
                growse.loadingNav = true;
                $.getJSON("/navlist/since/" + $("#articlenav > li:first").data('datestamp'), function (data) {
                    $.each(data, function (i, v) {
                        var markup = "<li><a href=\"/" + v.year + "/" + v.month + "/" + v.day + "/" + v.shorttitle + "/\" title=\"" + v.title + "\"<span>" + v.title + "</span></a></li>";
                        $("#articlenav").prepend(markup);
                    });
                });
                $.getJSON("/navlist/before/" + $("#articlenav > li:last").data('datestamp'), function (data) {
                    $.each(data, function (i, v) {
                        var markup = "<li><a href=\"/" + v.year + "/" + v.month + "/" + v.day + "/" + v.shorttitle + "/\" title=\"" + v.title + "\"<span>" + v.title + "</span></a></li>";
                        $("#articlenav").append(markup);
                    });

                });
            }
        },
        getLocation: function () {
            $.getJSON("//res.growse.com/nocache/latitude.js", function (data) {
                var coords = data.features[0].geometry.coordinates[1] + ',' + data.features[0].geometry.coordinates[0];
                var url = '//maps.googleapis.com/maps/api/staticmap?markers=color:red|' + coords + '&zoom=13&size=285x200&sensor=false';
                $('#twitterlocation_div p').html("<a href=\"//maps.google.com?q=" + coords + "\"><img src=" + url + " /></a>");
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
function prettyDate(time) {
    var date = new Date((time || "").replace(/-/g, "/").replace(/[TZ]/g, " ")),
        diff = (((new Date()).getTime() - date.getTime()) / 1000),
        day_diff = Math.floor(diff / 86400);

    if (isNaN(day_diff) || day_diff < 0 || day_diff >= 31)
        return;

    return day_diff == 0 && (
        diff < 60 && "just now" ||
            diff < 120 && "1 minute ago" ||
            diff < 3600 && Math.floor(diff / 60) + " minutes ago" ||
            diff < 7200 && "1 hour ago" ||
            diff < 86400 && Math.floor(diff / 3600) + " hours ago") ||
        day_diff == 1 && "Yesterday" ||
        day_diff < 7 && day_diff + " days ago" ||
        day_diff < 31 && Math.ceil(day_diff / 7) + " weeks ago";
}

// If jQuery is included in the page, adds a jQuery plugin to handle it as well
if (typeof jQuery != "undefined")
    jQuery.fn.prettyDate = function () {
        return this.each(function () {
            var date = prettyDate(this.title);
            if (date)
                jQuery(this).text(date);
        });
    };
