from django import template
from django.utils.safestring import mark_safe
import re

register = template.Library()


@register.filter(name='highlight')
def highlight(text, word):
    pattern = re.compile(r"(?P<filter>%s)" % word, re.IGNORECASE)
    return mark_safe(re.sub(pattern, r"<span class='highlight'>\g<filter></span>", text))