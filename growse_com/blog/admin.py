from django.forms import Textarea
from growse_com.blog.models import Article
from growse_com.blog.models import Comment
from django.contrib import admin
from django.db import models


class CommentInline(admin.TabularInline):
    model = Comment
    fields = ['datestamp', 'name', 'comment', 'ip']
    readonly_fields = ['ip']
    extra = 0


class ArticleAdmin(admin.ModelAdmin):
    fields = ['title', 'markdown']
    inlines = [CommentInline]
    list_display = ('title', 'datestamp')
    list_filter = ['datestamp']
    search_fields = ['title']
    formfield_overrides = {
        models.TextField: {'widget': Textarea(attrs={'class': 'input-xxlarge', 'rows': 20})},
    }


admin.site.register(Article, ArticleAdmin)
