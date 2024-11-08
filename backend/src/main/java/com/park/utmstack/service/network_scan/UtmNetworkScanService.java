package com.park.utmstack.service.network_scan;

import com.park.utmstack.domain.network_scan.NetworkScanFilter;
import com.park.utmstack.domain.network_scan.Property;
import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import com.park.utmstack.domain.network_scan.UtmPorts;
import com.park.utmstack.domain.network_scan.enums.AssetStatus;
import com.park.utmstack.domain.network_scan.enums.PropertyFilter;
import com.park.utmstack.domain.shared_types.enums.ImageShortName;
import com.park.utmstack.repository.UtmAssetMetricsRepository;
import com.park.utmstack.repository.UtmDataInputStatusRepository;
import com.park.utmstack.repository.network_scan.UtmNetworkScanRepository;
import com.park.utmstack.service.UtmImagesService;
import com.park.utmstack.service.agent_manager.AgentGrpcService;
import com.park.utmstack.service.dto.network_scan.NetworkScanDTO;
import com.park.utmstack.util.PdfUtil;
import com.park.utmstack.util.UtilPagination;
import com.park.utmstack.web.rest.errors.AgentNotfoundException;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.dao.InvalidDataAccessResourceUsageException;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.Assert;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import javax.persistence.EntityManager;
import java.io.ByteArrayOutputStream;
import java.util.*;
import java.util.stream.Collectors;

/**
 * Service Implementation for managing UtmNetworkScan.
 */
@Service
@Transactional
public class UtmNetworkScanService {

    private static final String CLASSNAME = "UtmNetworkScanService";
    private static final String NETWORK_SCAN_REPORT_TEMPLATE = "reports/network-scan/networkScanReport";

    private final UtmNetworkScanRepository networkScanRepository;
    private final UtmAssetMetricsRepository assetMetricsRepository;
    private final EntityManager em;
    private final PdfUtil pdfUtil;
    private final List<Long> newAssetIdRegistry = new ArrayList<>();
    private final UtmPortsService portsService;
    private final UtmImagesService imagesService;
    private final UtmDataInputStatusRepository utmDataInputStatusRepository;
    private final AgentGrpcService agentGrpcService;

    public UtmNetworkScanService(UtmNetworkScanRepository networkScanRepository,
                                 UtmAssetMetricsRepository assetMetricsRepository,
                                 EntityManager em, PdfUtil pdfUtil,
                                 UtmPortsService portsService, UtmImagesService imagesService,
                                 UtmDataInputStatusRepository utmDataInputStatusRepository,
                                 AgentGrpcService agentGrpcService) {
        this.networkScanRepository = networkScanRepository;
        this.assetMetricsRepository = assetMetricsRepository;
        this.em = em;
        this.pdfUtil = pdfUtil;
        this.portsService = portsService;
        this.imagesService = imagesService;
        this.utmDataInputStatusRepository = utmDataInputStatusRepository;
        this.agentGrpcService = agentGrpcService;
    }

    public void saveOrUpdateCustomAsset(NetworkScanDTO assetDto) throws Exception {
        final String ctx = CLASSNAME + ".saveOrUpdateCustomAsset";
        try {
            Assert.notNull(assetDto, "Asset information object is null");

            UtmNetworkScan nsAsset = networkScanRepository.save(new UtmNetworkScan(assetDto));
            final Long nsAssetId = nsAsset.getId();

            // Saving ports
            if (!CollectionUtils.isEmpty(assetDto.getPorts()))
                portsService.updateInBatch(nsAssetId, assetDto.getPorts().stream()
                    .map(port -> new UtmPorts(port, nsAssetId)).collect(Collectors.toList()));
            else portsService.deleteAllByScanId(nsAssetId);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Save a utmNetworkScan.
     *
     * @param utmNetworkScan the entity to save
     * @return the persisted entity
     */
    public UtmNetworkScan save(UtmNetworkScan utmNetworkScan) throws Exception {
        final String ctx = CLASSNAME + ".save";
        try {
            return networkScanRepository.save(utmNetworkScan);
        } catch (DataIntegrityViolationException e) {
            String msg = ctx + ": " + e.getMostSpecificCause().getMessage().replaceAll("\n", "");
            throw new RuntimeException(msg);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Save in batch a set of utmNetworkScan entity
     *
     * @param scans List of {@link UtmNetworkScan} to save
     * @throws Exception In case of any error
     */
    public void saveAll(List<UtmNetworkScan> scans) throws Exception {
        final String ctx = CLASSNAME + ".saveAll";
        try {
            networkScanRepository.saveAll(scans);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Update the type of each asset of the list
     *
     * @throws Exception In case of any error
     */
    public void updateType(List<Long> assetIds, Long assetTypeId) throws Exception {
        final String ctx = CLASSNAME + ".updateType";
        Assert.notEmpty(assetIds, ctx + ": Missing parameter [assetIds]");
        try {
            networkScanRepository.updateType(assetIds, assetTypeId);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public void updateGroup(List<Long> assetIds, Long assetGroupId) throws Exception {
        final String ctx = CLASSNAME + ".updateGroup";
        Assert.notEmpty(assetIds, ctx + ": Missing parameter [assetIds]");
        try {
            networkScanRepository.updateGroup(assetIds, assetGroupId);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Find all network scan results without paginating them
     *
     * @return List of {@link UtmNetworkScan}
     * @throws Exception In case of any error
     */
    @Transactional(readOnly = true)
    public List<UtmNetworkScan> findAll() throws Exception {
        final String ctx = CLASSNAME + ".findAll";
        try {
            return networkScanRepository.findAll();
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Get one utmNetworkScan by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    public Optional<NetworkScanDTO> searchDetails(Long id) throws Exception {
        final String ctx = CLASSNAME + ".searchDetails";
        try {
            return networkScanRepository.findById(id).map(m -> {
                m.setMetrics(assetMetricsRepository.findAllByAssetName(m.getAssetName()));
                return new NetworkScanDTO(m, true);
            });
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Gets network scan result by filters
     *
     * @param f Object with all filters to be applied
     * @param p For paginate the result
     * @return A list of {@link NetworkScanDTO} with results
     * @throws RuntimeException In case of any error
     */
    public Page<NetworkScanDTO> searchByFilters(NetworkScanFilter f, Pageable p) throws RuntimeException {
        final String ctx = CLASSNAME + ".searchByFilters";
        try {
            Page<UtmNetworkScan> page = filter(f, p);
            return page.map(m -> new NetworkScanDTO(m, false));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Search all no repeated values for a property field
     *
     * @param prop     Property field to get the values
     * @param pageable For paginate the result
     * @return A list with values of the property field
     * @throws Exception In case of any error
     */
    public List<?> searchPropertyValues(Property prop, String value, boolean forGroups, Pageable pageable) throws Exception {
        final String ctx = CLASSNAME + ".searchPropertyValues";
        try {
            StringBuilder sb = new StringBuilder("SELECT %1$s, COUNT(*) AS num FROM %2$s as p");

            if (!prop.getJoinTable().isEmpty()){
               sb.append(" %4$s");
            }

            sb.append(" WHERE (%1$s IS NOT NULL)");

            if (forGroups)
                sb.append(" AND (groupId IS NOT NULL)");

            if (StringUtils.hasText(value)) {
                if (prop == PropertyFilter.PORTS)
                    sb.append(" AND cast(port as string) LIKE '%%%3$s%%'");
                else if(prop == PropertyFilter.ALIVE && !value.isEmpty()){
                    sb.append(" AND %1$s = %3$s");
                }
                else
                    sb.append(" AND lower(%1$s) LIKE '%%%3$s%%'");
            }
            sb.append(" GROUP BY %1$s");

            String query = String.format(sb.toString(), prop.getPropertyName(), prop.getFromTable(), StringUtils.hasText(value) ? value.toLowerCase() : null, prop.getJoinTable());

            return em.createQuery(query).setFirstResult(UtilPagination.getFirstForNativeSql(pageable.getPageSize(), pageable.getPageNumber())).setMaxResults(
                pageable.getPageSize()).getResultList();
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Count all new assets discovered
     *
     * @return A number with amount of new discovered assets
     * @throws Exception In case of any error
     */
    public Integer countNewAssets() throws Exception {
        final String ctx = CLASSNAME + ".countNewAssets";

        try {
            Optional<List<UtmNetworkScan>> newAssets = networkScanRepository.findAllByAssetStatus(AssetStatus.NEW);
            int counter = 0;

            if (newAssets.isEmpty())
                return 0;

            List<UtmNetworkScan> assets = newAssets.get();

            for (UtmNetworkScan asset : assets) {
                if (!newAssetIdRegistry.contains(asset.getId())) {
                    counter++;
                    newAssetIdRegistry.add(asset.getId());
                }
            }
            return counter;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public ByteArrayOutputStream getNetworkScanReport(NetworkScanFilter f, Pageable p) throws Exception {
        final String ctx = CLASSNAME + ".getNetworkScanReport";
        try {
            Page<UtmNetworkScan> page = filter(f, p);

            if (page.getTotalPages() <= 0)
                return null;

            List<UtmNetworkScan> assets = page.getContent();
            Map<String, Object> vars = new HashMap<>();
            vars.put("assets", assets);
            vars.put("logo", imagesService.findOne(ImageShortName.REPORT).orElse(null));
            return pdfUtil.convertHtmlTemplateToPdf(NETWORK_SCAN_REPORT_TEMPLATE, vars);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    private Page<UtmNetworkScan> filter(NetworkScanFilter f, Pageable p) throws Exception {
        final String ctx = CLASSNAME + ".filter";
        try {
            Page<UtmNetworkScan> page = networkScanRepository.searchByFilters(
                f.getAssetIpMacName() == null ? null : "%" + f.getAssetIpMacName() + "%",
                f.getOs(), f.getAlias(), f.getType(), f.getAlive(), f.getStatus(),
                f.getProbe(), f.getOpenPorts(), f.getDiscoveredInitDate(),
                f.getDiscoveredEndDate(), f.getGroups(), f.getRegisteredMode(), f.getAgent(), f.getOsPlatform(), f.getDataTypes(), p);

            /*if (page.getTotalPages() > 0) {
                List<UtmDataInputStatus> utmDataInputStatuses = utmDataInputStatusRepository.findAll().stream().sorted(Comparator.comparing(UtmDataInputStatus::getSource)).collect(Collectors.toList());
                page.forEach(m -> m.setMetrics(assetMetricsRepository.findAllByAssetName(m.getAssetName())));
                page.forEach(m -> m.setDataInputList(utmDataInputStatuses.stream().filter(
                    inputStatus -> inputStatus.getSource().equalsIgnoreCase(m.getAssetName()) ||
                        inputStatus.getSource().equalsIgnoreCase(m.getAssetIp())).collect(Collectors.toList())));
            }*/

            return page;
        } catch (InvalidDataAccessResourceUsageException e) {
            String msg = ctx + ": " + e.getMostSpecificCause().getMessage().replaceAll("\n", "");
            throw new Exception(msg);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Remove the asset information and all his data sources. Also return source information if it is an agent,
     * that information will be used to remove agent from agent manager
     *
     * @param id Asset identifier
     */
    public void deleteCustomAsset(Long id) {
        final String ctx = CLASSNAME + ".deleteCustomAsset";
        String src = "";
        Optional<UtmNetworkScan> asset = networkScanRepository.findById(id);
        try {
            if (asset.isPresent()) {
                src = StringUtils.hasText(asset.get().getAssetName()) ? asset.get().getAssetName() : asset.get().getAssetIp();

                if (asset.get().getIsAgent())
                    agentGrpcService.deleteAgent(src);
                utmDataInputStatusRepository.deleteAllBySource(src);
                networkScanRepository.deleteById(id);
            };
        } catch (AgentNotfoundException e) {
            // If agent wasn't found, remove from database anyway
            if (StringUtils.hasText(src)) {
                utmDataInputStatusRepository.deleteAllBySource(src);
                networkScanRepository.deleteById(id);
            }
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Checks if the asset compliance the necessary requirement to run a command over it.<br/>
     * Requirements are:<br/>
     * <li>The asset has to exist</li>
     * <li>The asset has to be an agent</li>
     * <li>The asset can't have the MISSING status</li>
     * <li>The asset has to be active</li>
     *
     * @param assetName Name of the asset
     * @return true if all necessary requirements to execute a command are met
     * @throws Exception In case of any error
     */
    public boolean canRunCommand(String assetName) throws Exception {
        final String ctx = CLASSNAME + ".canRunCommand";
        try {
            if (!StringUtils.hasText(assetName))
                throw new Exception("Parameter assetName is invalid");
            Optional<UtmNetworkScan> assetOpt = networkScanRepository.findByAssetName(assetName);
            if (assetOpt.isEmpty())
                return false;
            UtmNetworkScan asset = assetOpt.get();
            return asset.getIsAgent() && !asset.getAssetStatus().equals(AssetStatus.MISSING) && asset.getAssetAlive();
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public Optional<UtmNetworkScan> findByNameOrIp(String nameOrIp) {
        final String ctx = CLASSNAME + ".findByNameOrIp";
        try {
            if (!StringUtils.hasText(nameOrIp))
                throw new Exception("Parameter nameOrIp is null or empty");
            return networkScanRepository.findByNameOrIp(nameOrIp);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public List<String> getAgentsOsPlatform() {
        final String ctx = CLASSNAME + ".getAgentsOsPlatform";
        try {
            return networkScanRepository.findAgentsOsPlatform();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }
}
