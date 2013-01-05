from growse_com.blog.models import Article
from growse_com.blog.models import Comment
from growse_com.blog.models import Footerline
from django.contrib import admin


class CommentInline(admin.TabularInline):
    model = Comment
    fields = ['datestamp', 'name', 'comment', 'ip']
    readonly_fields = ['ip']
    extra = 0


class ArticleAdmin(admin.ModelAdmin):
    fields = ['datestamp', 'title', 'markdown', 'type']
    inlines = [CommentInline]
    list_display = ('title', 'datestamp')
    list_filter = ['datestamp']
    search_fields = ['title']
#    date_hierarchy = 'datestamp'


admin.site.register(Article, ArticleAdmin)
admin.site.register(Footerline)
