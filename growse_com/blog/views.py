from django.core.mail import send_mail
from django.template import RequestContext
from django.shortcuts import get_object_or_404, redirect, render_to_response
from django.db.models import Count
from django.http import Http404
from django.core.paginator import Paginator, InvalidPage, EmptyPage
from growse_com.blog.models import Article
from growse_com.blog.models import Comment

def article(request, article_shorttitle=''):
    c = RequestContext(request)
    if article_shorttitle == '':
        article = Article.objects.latest('datestamp')
    else:
        article = get_object_or_404(Article, shorttitle=article_shorttitle)
    if request.method == 'POST':
        name = request.POST.get('name')
        website = request.POST.get('website')
        comment = request.POST.get('comment')
        spamfilter = request.POST.get('email')
        if spamfilter is None or len(spamfilter) == 0:
            Comment.objects.create(name=name, website=website, comment=comment, article=article, ip=request.META['REMOTE_ADDR'])
            try:
                send_mail('New Comment on growse.com', 'Someone posted a comment on growse.com. Over at http://www.growse.com/news/comments/' + article.shorttitle + '/', 'hubfour@growse.com', ['comments@growse.com'], fail_silently=False)
            except:
                pass
        return redirect("/news/comments/" + article_shorttitle + "/")
    else:
        articlenavlist = Article.objects.all().order_by('-datestamp')
        prevarticle = None
        nextarticle = None
        if article.type == 'NEWS':
            prevarticle = Article.objects.filter(type='NEWS').filter(datestamp__lt=article.datestamp).order_by("-datestamp")[0:1]
            nextarticle = Article.objects.filter(type='NEWS').filter(datestamp__gt=article.datestamp).order_by("datestamp")[0:1]
            if prevarticle.count() != 0:
                prevarticle = prevarticle[0]
            if nextarticle.count() != 0:
                nextarticle = nextarticle[0]
        comments = Comment.objects.filter(article__id=article.id).order_by("datestamp")
        archives = Article.objects.filter(type='NEWS').extra(select={'month': "DATE_TRUNC('month',datestamp)"}).values('month').annotate(Count('title')).order_by('-month')
        prevyear = None
        for archive in archives:
            if archive["month"].year != prevyear:
                archive["newyear"] = True
                prevyear = archive["month"].year
        return render_to_response('blog/article.html', {'archives': archives, 'articlenavlist': articlenavlist, 'comments': comments, 'article': article, 'nav': article.type.lower()}, c)


def search(request, searchterm=None, page=1):
    c = RequestContext(request)
    if searchterm is None:
        if request.method == 'GET':
            return redirect("/", Permanent=True)
        if request.method == 'POST':
            return redirect("/search/" + request.POST.get('a','') + "/")
    else:
        results_list =  Article.objects.extra(select={'headline':"ts_headline(body,plainto_tsquery('english',%s))",'rank':"ts_rank(idxfti,plainto_tsquery('english',%s))"},where=["idxfti @@ plainto_tsquery('english',%s)"],params=[searchterm],select_params=[searchterm,searchterm]).order_by('-rank')
        paginator = Paginator(results_list,10)
        try:
            results = paginator.page(page)
        except(EmptyPage, InvalidPage):
            results = paginator.page(paginator.num_pages)
        return render_to_response('blog/search.html',{'results':results,'searchterm':searchterm}, c)

def links (request):
    c=RequestContext(request);
    return render_to_response('blog/links.html',{'nav':'links'},c)

