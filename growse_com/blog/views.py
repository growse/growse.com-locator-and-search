from django.core.mail import send_mail
from django.template import RequestContext
from django.shortcuts import get_object_or_404, redirect, render_to_response
from django.db.models import Count
from django.http import HttpResponsePermanentRedirect, HttpResponse, Http404
from django.core.paginator import Paginator, InvalidPage, EmptyPage
from growse_com.blog.models import Article
from growse_com.blog.models import Comment
import simplejson as json


def article_shorttitle(request, article_shorttitle=''):
    article = get_object_or_404(Article, shorttitle=article_shorttitle)
    articledate = article.datestamp.date()
    return HttpResponsePermanentRedirect(
        '/' + str(articledate.year) + '/' + str(articledate.month).zfill(2) + '/' + str(
            articledate.day).zfill(2) + '/' + article.shorttitle + '/')


def article_bydate(request, year, month='', day=''):
    article = None
    if day and month and year:
        try:
            article = Article.objects.filter(datestamp__year=year, datestamp__month=month, datestamp__day=day).order_by(
                'datestamp')[0]
        except IndexError:
            raise Http404
    elif month and year:
        try:
            article = Article.objects.filter(datestamp__year=year, datestamp__month=month).order_by('datestamp')[0]
        except IndexError:
            raise Http404
    elif year:
        try:
            article = Article.objects.filter(datestamp__year=year).order_by('datestamp')[0]
        except IndexError:
            raise Http404

    if article:
        articledate = article.datestamp.date()
        return redirect('/' + str(articledate.year) + '/' + str(articledate.month).zfill(2) + '/' + str(
            articledate.day).zfill(2) + '/' + article.shorttitle + '/')


def navlist(request, direction, datestamp):
    if direction == 'before':
        articles = Article.objects.filter(datestamp__lt=datestamp).order_by('-datestamp')
    elif direction == 'since':
        articles = Article.objects.filter(datestamp__gt=datestamp).order_by('-datestamp')
    response_data = []
    for article in articles:
        response_data.append({
            'title': article.title,
            'id': article.id,
            'shorttitle': article.shorttitle,
            'year': str(article.datestamp.year).zfill(4) if article.datestamp else None,
            'month': str(article.datestamp.month).zfill(2) if article.datestamp else None,
            'day': str(article.datestamp.day).zfill(2) if article.datestamp else None
        })
    return HttpResponse(json.dumps(response_data), content_type='application/json')


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
        articledate = article.datestamp.date()
        if spamfilter is None or len(spamfilter) == 0:
            Comment.objects.create(name=name, website=website, comment=comment, article=article,
                                   ip=request.META['REMOTE_ADDR'])
            try:
                send_mail('New Comment on growse.com',
                          'Someone posted a comment on growse.com. Over at http://www.growse.com/' + str(
                              articledate.year) + '/' + str(articledate.month).zfill(2) + '/' + str(
                              articledate.day).zfill(2) + '/' + article.shorttitle + '/',
                          'hubfour@growse.com', ['comments@growse.com'], fail_silently=False)
            except:
                pass
        return redirect('/' + str(articledate.year) + '/' + str(articledate.month).zfill(2) + '/' + str(
            articledate.day).zfill(2) + '/' + article.shorttitle + '/')
    else:
        navitems = Article.objects.raw(
            "(select id,title,datestamp,shorttitle from articles where id=%(id)s)"
            " union"
            " (select id,title,datestamp,shorttitle from articles where datestamp<(select datestamp from articles where id=%(id)s) order by datestamp desc limit 10)"
            " union"
            " (select id,title,datestamp,shorttitle from articles where datestamp>(select datestamp from articles where id=%(id)s) order by datestamp asc limit 10) order by datestamp asc;",
            {'id': article.id}
        )

        comments = Comment.objects.filter(article__id=article.id).order_by("datestamp")
        archives = Article.objects.filter(type='NEWS').extra(select={'month': "DATE_TRUNC('month',datestamp)"}).values(
            'month').annotate(Count('title')).order_by('-month')
        prevyear = None
        for archive in archives:
            if archive["month"].year != prevyear:
                archive["newyear"] = True
                prevyear = archive["month"].year
        return render_to_response('article.html',
                                  {'archives': archives, 'navitems': navitems, 'comments': comments,
                                   'article': article}, c)


def search(request, searchterm=None, page=1):
    c = RequestContext(request)
    if searchterm is None:
        if request.method == 'GET':
            return redirect("/", Permanent=True)
        if request.method == 'POST':
            return redirect("/search/" + request.POST.get('a', '') + "/")
    else:
        results_list = Article.objects.extra(select={
            'headline': "ts_headline(body,plainto_tsquery('english',%s),'MaxFragments=1, MinWords=5, MaxWords=25')",
            'rank': "ts_rank(idxfti,plainto_tsquery('english',%s))"},
                                             where=["idxfti @@ plainto_tsquery('english',%s)"], params=[searchterm],
                                             select_params=[searchterm, searchterm]).order_by('-rank')
        paginator = Paginator(results_list, 10)
        try:
            results = paginator.page(page)
        except(EmptyPage, InvalidPage):
            results = paginator.page(paginator.num_pages)
        return render_to_response('search.html', {'results': results, 'searchterm': searchterm}, c)


