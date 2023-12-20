package com.park.utmstack.service.reports;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import com.park.utmstack.domain.index_pattern.enums.SystemIndexPattern;
import com.park.utmstack.domain.network_scan.NetworkScanFilter;
import com.park.utmstack.domain.reports.types.IncidentType;
import com.park.utmstack.domain.shared_types.AlertType;
import com.park.utmstack.domain.shared_types.enums.ImageShortName;
import com.park.utmstack.service.UtmImagesService;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.service.elasticsearch.SearchUtil;
import com.park.utmstack.service.incident_response.UtmIncidentJobService;
import com.park.utmstack.service.network_scan.UtmNetworkScanService;
import com.park.utmstack.util.PdfUtil;
import com.park.utmstack.util.enums.AlertStatus;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.opensearch.client.opensearch.core.search.Hit;
import org.opensearch.client.opensearch.core.search.HitsMetadata;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.io.ByteArrayOutputStream;
import java.time.Instant;
import java.util.*;
import java.util.stream.Collectors;

@Service
public class CustomReportService {
    private static final String CLASSNAME = "CustomReportService";

    private final ElasticsearchService elasticsearchService;
    private final PdfUtil pdfUtil;
    private final UtmIncidentJobService incidentJobService;
    private final UtmNetworkScanService utmNetworkScanService;
    private final UtmImagesService imagesService;

    public CustomReportService(ElasticsearchService elasticsearchService,
                               PdfUtil pdfUtil,
                               UtmIncidentJobService incidentJobService,
                               UtmNetworkScanService utmNetworkScanService,
                               UtmImagesService imagesService) {
        this.elasticsearchService = elasticsearchService;
        this.pdfUtil = pdfUtil;
        this.incidentJobService = incidentJobService;
        this.utmNetworkScanService = utmNetworkScanService;
        this.imagesService = imagesService;
    }

    /**
     * Build a threat alert activity report
     *
     * @param from Initial date range
     * @param to   End of the date range
     * @param top  Result top
     * @return A pdf with report information
     * @throws Exception In case of any error
     */
    public Optional<ByteArrayOutputStream> buildThreatActivityForAlerts(Instant from, Instant to, Integer top) throws Exception {
        final String ctx = CLASSNAME + ".buildThreatActivityForAlerts";
        try {
            List<FilterType> filters = new ArrayList<>(Arrays.asList(
                new FilterType(Constants.timestamp, OperatorType.IS_BETWEEN, Arrays.asList(from, to)),
                new FilterType(Constants.alertIsIncident, OperatorType.IS, false),
                new FilterType(Constants.alertStatus, OperatorType.IS_NOT, AlertStatus.AUTOMATIC_REVIEW.getCode())));

            Pageable page = PageRequest.of(0, top, Sort.by(Sort.Direction.ASC, Constants.alertStatus));

            SearchRequest.Builder srb = new SearchRequest.Builder();
            srb.index(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS))
                .query(SearchUtil.toQuery(filters));
            SearchUtil.applyPaginationAndSort(srb, page, top);

            HitsMetadata<AlertType> hits = elasticsearchService.search(srb.build(), AlertType.class).hits();

            if (hits.total().value() <= 0)
                return Optional.empty();

            List<AlertType> alerts = hits.hits().stream().map(Hit::source).collect(Collectors.toList());

            Map<String, Object> vars = new HashMap<>();
            vars.put("alerts", alerts);
            vars.put("logo", imagesService.findOne(ImageShortName.REPORT).orElse(null));

            return Optional.ofNullable(pdfUtil.convertHtmlTemplateToPdf("reports/customs/threatActivityForAlerts", vars));
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public Optional<ByteArrayOutputStream> buildThreatActivityForIncidents(Instant from, Instant to, Integer top) {
        final String ctx = CLASSNAME + ".buildThreatActivityForIncidents";
        try {
            List<FilterType> filters = new ArrayList<>(Arrays.asList(
                new FilterType(Constants.timestamp, OperatorType.IS_BETWEEN, Arrays.asList(from, to)),
                new FilterType(Constants.alertIsIncident, OperatorType.IS, true)));

            Pageable page = PageRequest.of(0, top, Sort.by(Sort.Direction.ASC, Constants.alertStatus));

            SearchRequest.Builder srb = new SearchRequest.Builder();
            srb.index(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS))
                .query(SearchUtil.toQuery(filters));
            SearchUtil.applyPaginationAndSort(srb, page, top);

            HitsMetadata<AlertType> hits = elasticsearchService.search(srb.build(), AlertType.class).hits();

            if (hits.total().value() <= 0)
                return Optional.empty();

            List<AlertType> incidentDocs = hits.hits().stream().map(Hit::source).collect(Collectors.toList());

            List<IncidentType> incidents = new ArrayList<>();

            for (AlertType incident : incidentDocs) {
                IncidentType incidentType = new IncidentType();
                incidentType.setIncident(incident);

                String src = "", dest = "";

                if (!Objects.isNull(incident.getSource()))
                    src = StringUtils.hasText(incident.getSource().getHost())
                        ? incident.getSource().getHost() : incident.getSource().getIp();

                if (!Objects.isNull(incident.getDestination()))
                    dest = StringUtils.hasText(incident.getDestination().getHost())
                        ? incident.getDestination().getHost() : incident.getDestination().getIp();

                if (StringUtils.hasText(src))
                    incidentType.setSrcResponses(incidentJobService.findAllByAgent(src));

                if (StringUtils.hasText(dest))
                    incidentType.setDestResponses(incidentJobService.findAllByAgent(dest));

                incidents.add(incidentType);
            }

            Map<String, Object> vars = new HashMap<>();
            vars.put("incidents", incidents);
            vars.put("logo", imagesService.findOne(ImageShortName.REPORT).orElse(null));

            return Optional.ofNullable(pdfUtil.convertHtmlTemplateToPdf("reports/customs/threatActivityForIncidents", vars));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public Optional<ByteArrayOutputStream> buildAssetManagement(Instant from, Instant to, Integer top) {
        final String ctx = CLASSNAME + ".buildAssetManagement";

        try {
            NetworkScanFilter filters = new NetworkScanFilter();
            filters.setDiscoveredInitDate(from);
            filters.setDiscoveredEndDate(to);

            Pageable page = PageRequest.of(0, top, Sort.by(Sort.Direction.DESC, "discoveredAt"));
            return Optional.ofNullable(utmNetworkScanService.getNetworkScanReport(filters, page));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }
}
