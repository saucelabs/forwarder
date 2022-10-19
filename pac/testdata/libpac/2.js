// This PAC file is from http://cns.ntou.edu.tw/lib.pac.

function FindProxyForURL(url, host) {
var RESOLV_IP;
var lchost = host.toLowerCase();
if(check(host,"*.*.ebscohost.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"*.ebsco-content.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"129.35.213.31",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"129.35.248.48",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"134.243.85.3",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"134.243.85.4",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"140.121.140.100",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"140.121.140.102",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"140.121.140.103",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"140.121.180.109",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"156.csis.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"165.193.122.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"165.193.141.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"167.216.170.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"167.216.171.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"170.225.184.106",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"170.225.184.107",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"170.225.96.21",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"170.225.99.9",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"192.83.186.103",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"192.83.186.70",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"192.83.186.71",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"192.83.186.72",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"192.83.186.84",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"199.4.154.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"199.4.155.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"202.70.173.2",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"203.70.208.88",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"203.74.36.75",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"205.240.244.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"205.240.245.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"205.240.246.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"205.240.247.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"205.243.231.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"210.243.166.93",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"211.20.182.42",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"211.79.206.2",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"211.79.206.4",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"211.79.506.4",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"220.228.59.156",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"63.240.105.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"63.240.113.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"63.84.162.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"63.86.118.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"63.86.119.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"65.246.184.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"65.246.185.",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"aac.asm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ac.els-cdn.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"admin-apps.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"admin-router.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"admin.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"aem.asm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"afraf.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ageing.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"aje.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"alcalc.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"aler.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"annhyg.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"annonc.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"antonio.ingentaselect.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ao.osa.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"aob.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"aoip.osa.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"aolp.osa.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"aoot.osa.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ap.ejournal.ascc.net",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"apl.aip.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"apollo.sinica.edu.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"apps.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"apps.webofknowledgev4.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"arjournals.annualreviews.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ascelibrary.aip.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"atoz.ebsco.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"beck-online.beck.de",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"beheco.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"bencao.infolinker.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"big5.oversea.cnki.net",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"bioinformatics.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"biostatistics.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"bizboard.nikkeibp.co.jp",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"bizboard.nikkeibp.co.jp/daigaku",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"bja.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"bjc.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"bjsw.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"bmb.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"bmf.aip.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"brain.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"brief-treatment.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"carcin.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cco.cambridge.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cdj.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cdli.asm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cds1.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cds2.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cec.lib.apabi.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cep.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cercor.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"chaos.aip.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"charts.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"chemse.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ci.nii.ac.jp",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cje.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cjn.csis.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"clipsy.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cm.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cm.webofknowledgev4.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cmr.asm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cnki.csis.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cnki50.csis.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"collections.chadwyck.co.uk",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"concert.wisenews.net.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"content.ebscohost.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cornell.mirror.aps.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"cpe.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"csa.e-lib.nctu.edu.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ct.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"database.yomiuri.co.jp",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"db.lib.ntou.edu.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"deafed.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"delivery.acm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"demomars.csis.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"diipcs.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"dlib.apabi.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"download.springer.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ea.grolier.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"earthinteractions.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ebook01.koobe.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ebooks.abc-clio.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ebooks.kluweronline.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ebooks.springerlink.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ebooks.windeal.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ec.asm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"edo.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"edo.tw/ocp.aspx?subs_no=20063",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"eds.a.ebscohost.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"eds.b.ebscohost.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"eds.c.ebscohost.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"eds.d.ebscohost.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"eds.e.ebscohost.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"edu1.wordpedia.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"eebo.chadwyck.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ei.e-lib.nctu.edu.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ei.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ei.stic.gov.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ej.iop.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"elearning.webenglish.tv",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"elib.infolinker.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"emboj.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"engineer.windeal.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"enterprise.astm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"epirev.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"epubs.siam.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"erae.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"eric.lib.nccu.edu.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"es.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"esi.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"esr.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"estipub.isiknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ethesys.lib.ntou.edu.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"fampra.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"g.wanfangdata.com.hk",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"galenet.galegroup.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"gateway.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"german2.nccu.edu.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"global.ebsco-content.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"global.umi.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"globalbb.onesource.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"glycob.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"gme.grolier.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"go-passport.grolier.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"go.galegroup.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"go.grolier.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"go.westlawjapan.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"haworthpress.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"hbrtwn.infolinker.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"hcr.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"hcr3.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"heapol.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"heapro.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"her.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"hjournals.cambridge.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"hk.wanfangdata.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"hmg.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"http://infotrac.galegroup.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"humrep.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"hunteq.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"huso.stpi.narl.org.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"huso.stpi.org.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"hyweb.ebook.hyread.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"iai.asm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"icc.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ieee.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"igroup.ebrary.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ije.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ijpor.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ilibrary.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"imagebank.osa.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"images.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"infotrac.galegroup.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"infoweb.newsbank.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"international.westlaw.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"intimm.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"intqhc.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"iopscience.iop.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"iospress.metapress.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"irap.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"isi4.isiknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"isiknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jac.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jae.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jap.aip.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jb.asm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jcm.asm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jcp.aip.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jcr1.isiknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jeg.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jhered.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jjco.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jleo.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jlt.osa.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jmicro.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jmp.aip.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jn.physiology.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jncicancerspectrum.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"joc.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"josaa.osa.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jot.osa.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"journals.ametsoc.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"journals.asm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"journals.cambridge.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"journals.kluweronline.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"journals.wspc.com.sg",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jpart.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jpcrd.aip.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jpepsy.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jrse.aip.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jurban.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jvi.asm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"jxb.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"klassiker.chadwyck.co.uk",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"kmw.ctgin.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"lang.ntou.edu.tw/source.php",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"lb20.ah100.libraryandbook.net",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"lb20.botw.libraryandbook.net",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"lb20.dummies.libraryandbook.net",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"lb20.tabf.libraryandbook.net",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"legal.lexisnexis.jp",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"lib.myilibrary.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"library.books24x7.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"library.pressdisplay.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"link.aps.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"link.springer-ny.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"link.springer.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"link.springer.de",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"links.springer.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"links.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ltp.aip.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"mars.csa.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"mars.csis.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"mars2.csa.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"mars3.csa.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"mbe.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"mcb.asm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"md1.csa.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"md2.csa.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"minghouse.infolinker.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"mmbr.asm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"molehr.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"mollus.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"mutage.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"mydigitallibrary.lib.overdrive.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"nar.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ncl3web.hyweb.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ndt.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"new.cwk.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"newfirstsearch.global.oclc.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"newfirstsearch.oclc.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ntou.ebook.hyread.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ntou.koobe.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ntt1.hyweb.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"occmed.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"oep.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"oh1.csa.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"oh2.csa.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ojps.aip.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ol.osa.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"oldweb.cqvip.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"omed.nuazure.info",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"online.sagepub.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"onlinelibrary.wiley.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ortho.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"oversea.cnki.net",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ovid.stic.gov.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ovidsp.ovid.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"oxfordjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"oxrep.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pa.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pan.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pao.chadwyck.co.uk",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pcp.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pcs.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pcs.webofknowledgev4com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pdn.sciencedirect.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"petrology.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"phr.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"physics.aps.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"physiolgenomics.physiology.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"plankt.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pm.nlx.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pm.nlx.com/xtf/search?browse-collections=true",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pof.aip.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pop.aip.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"portal.acm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"portal.isiknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pqdd.sinica.edu.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pra.aps.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"prb.aps.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"prc.aps.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"prd.aps.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pre.aps.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"prl.aps.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pro-twfubao.infolinker.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"prola.aps.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"prola.library.cornell.edu",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"proquest.umi.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"proquest.uni.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"protein.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ptr.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pubmed.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pubs.acs.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pubs.rsc.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"pubs3.acs.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"qjmed.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"reading.udn.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"readopac.ncl.edu.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"readopac2.ncl.edu.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"readopac3.ncl.edu.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"reference.kluweronline.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"refworks.reference-global.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"rfs.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"rheumatology.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"rmp.aps.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"rsi.aip.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"rss.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"schiller.chadwyck.co.uk",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"sciencenow.sciencemag.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"scifinder.cas.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"scitation.aip.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"sdos.ejournal.ascc.net",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"search.ebscohost.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"search.epnet.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"search.isiknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"search.proquest.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"search.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ser.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"service.csa.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"service.flysheet.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"service.refworks.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"shmu.alexanderstreet.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"site.ebrary.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"soth.alexanderstreet.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"sp.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"springerlink.metapress.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ssjj.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"stfb.ntl.edu.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"stfj.ntl.edu.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"sub3.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"survival.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"sushi.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"swproxy.swetswise.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"taebc.ebook.hyread.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"taebc.etailer.dpsl.net",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"taebc.koobe.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"taebcmgh.sa.libraryandbook.net",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"tandf.msgfocus.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"tao.wordpedia.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"tbmcdb.infolinker.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"tcsd.lib.ntu.edu.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"tebko.infolinker.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"tie.tier.org.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"toc.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"toc.webofknowledgev4.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"tongji.oversea.cnki.net",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"toxsci.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"turs.infolinker.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"tw.magv.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"twu-ind.wisenews.net.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"udndata.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"udndata.com/library/fullpage",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ulej.stic.gov.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"vnweb.hwwilsonweb.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"wber.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"wbro.oupjournals.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"wcs.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"web.a.ebscohost.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"web.b.ebscohost.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"web.c.ebscohost.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"web.d.ebscohost.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"web.e.ebscohost.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"web.ebscohost.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"web.lexis-nexis.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"web17.epnet.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"webofknowledge.com&nbsp;",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"wok-ws.isiknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"wos.stic.gov.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"ws.isiknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.acm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.agu.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.airitiaci.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.airitiaci.com/",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.airitibooks.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.airitilibrary.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.airitinature.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.annualreviews.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.apabi.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.apabi.com/cec",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.ascelibrary.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.asme.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.astm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.atozmapsonline.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.atoztheworld.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.atozworldbusiness.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.atozworldculture.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.atozworldtrade.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.atozworldtravel.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.biolbull.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.bioone.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.blackwell-synergy.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.brepolis.net",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.bridgemaneducation.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.bssaonline.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.cairn.info",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.ceps.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.chinamaxx.net",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.classiques-garnier.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.cnsonline.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.cnsppa.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.crcnetbase.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.credoreference.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.csa.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.csa.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.dalloz.fr",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.dialogselect.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.discoverygate.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.duxiu.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.ebookstore.tandf.co.uk",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.ebsco.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.educationarena.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.ei.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.els.net",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.elsevier.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.emeraldinsight.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.engineeringvillage.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.engineeringvillage2.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.europaworld.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.europe.idealibrary.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.frantext.fr",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.genome.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.genomebiology.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.greeninfoonline.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.hepseu.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.icevirtuallibrary.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.idealibrary.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.igpublish.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.informaworld.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.ingenta.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.int-res.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.iop.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.iospress.nl",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.isihighlycited.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.jkn21.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.jstor.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.juris.de",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.kluwerlawonline.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.kluweronline.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.knovel.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.lawbank.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.lawdata.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.lexisnexis.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.lexisnexis.com/ap/academic",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.lexisnexis.com/ap/auth",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.lextenso.fr",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.mergentonline.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.mrw.interscience.wiley.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.munzinger.de",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.myendnoteweb.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.myilibrary.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.nature.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.netlibrary.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.nonlin-processes-geophys.net",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.nutrition.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.onesource.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.opticsexpress.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.osa-jon.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.osa-opn.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.oxfordreference.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.oxfordscholarship.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.palgrave-journals.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.palgraveconnect.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.proteinscience.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.read.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.reaxys.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.refworks.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.refworks.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.researcherid.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.rsc.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.sage-ereference.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.sciencedirect.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.sciencemag.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.scopus.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.springerlink.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.swetsnet.nl",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.swetswise.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.taebcnetbase.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.tandf.co.uk",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.tandfonline.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.tbmc.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.TeacherReference.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.tlemea.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.tls.psmedia.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.tumblebooks.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.tw-elsevier.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.universalis-edu.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.wanfangdata.com.hk",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.webofknowledge.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.webofknowledgev4.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.westlaw.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www.wkap.nl",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www2.astm.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www2.read.com.tw",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www3.electrochem.org",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www3.interscience.wiley.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"www3.oup.co.uk",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"wwwlib.global.umi.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
if(check(host,"yagi.jkn21.com",false,true))
        return "PROXY proxylib.ntou.edu.tw:3128";
return	"DIRECT";
}
function check(target,term,caseSens,wordOnly) {
if (!caseSens) {
term = term.toLowerCase();
target = target.toLowerCase();
}
if(target.indexOf(term) >= 0) {
return true;
}
return false;
}
