var growse = function () {
    return {
        loadNav: function () {
            $.getJSON("/navlist/", function (data) {
                var hereid = $("#articlenav > li").data('id');
                var before = true;
                $.each(data, function (i, v) {
                    if (v.id == hereid) {
                        before = false;
                        return;
                    }
                    var markup = "<li><a href=\"/" + v.year + "/" + v.month + "/" + v.day + "/" + v.shorttitle + "/\" title=\"" + v.title + "\"<span>" + v.title + "</span></a></li>";
                    if (before) {
                        $("#articlenav").prepend(markup);
                    } else {
                        $("#articlenav").append(markup);
                    }
                });
            });
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
