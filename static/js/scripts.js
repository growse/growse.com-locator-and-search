$(function () {
    var percentagedown = ($('.here').position().top / $(window).height()) * 100;
    if (percentagedown > 50) {
        var value = $('.here').position().top - ($(window).height() / 2) + ($('nav ul li:first').height() / 2);
        $(".nano").nanoScroller({scrollTop: value});
    } else {
        $(".nano").nanoScroller();
    }
});