import markdown

from django.template.defaultfilters import stringfilter
from django.utils.encoding import force_unicode
from django import template
from django.utils.safestring import mark_safe
import re

register = template.Library()


@register.filter(name='highlight')
def highlight(text, word):
    pattern = re.compile(r"(?P<filter>%s)" % word, re.IGNORECASE)
    return mark_safe(re.sub(pattern, r"<span class='highlight'>\g<filter></span>", text))



@register.filter(is_safe=True)
@stringfilter
def my_markdown(value):

    extensions = ["codehilite", "sane_lists"]
    return mark_safe(markdown.markdown(force_unicode(value),
                                       extensions,
                                       lazy_ol=False,
                                       safe_mode=False,
                                       enable_attributes=False))