<div class="footer">
    <div class="container">
        <div class="row text-center border-top">
            <span><a href="https://www.baidu.com" target="_blank">{{i18n .Lang "common.official_website"}}</a></span>
            <span>&nbsp;·&nbsp;</span>
            <span><a href="mailto:runzii@foxmail.com" target="_blank">{{i18n .Lang "common.feedback"}}</a></span>
            <span>&nbsp;·&nbsp;</span>
            <span><a href="https://www.baidu.com" target="_blank">{{i18n .Lang "common.manual"}}</a></span>
        </div>
        {{if .site_beian}}
        <div class="row text-center">
            <a href="https://beian.miit.gov.cn/" target="_blank">{{.site_beian}}</a>
        </div>
        {{end}}
    </div>
</div>
