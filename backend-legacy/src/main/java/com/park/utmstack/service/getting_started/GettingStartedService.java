package com.park.utmstack.service.getting_started;

import com.park.utmstack.domain.getting_started.GettingStartedStepEnum;
import com.park.utmstack.domain.getting_started.UtmGettingStarted;
import com.park.utmstack.repository.getting_started.UtmGettingStartedRepository;
import com.park.utmstack.service.application_modules.UtmModuleService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.List;

@Service
public class GettingStartedService {
    private final Logger log = LoggerFactory.getLogger(UtmModuleService.class);
    private static final String CLASSNAME = "GettingStartedService";

    private final UtmGettingStartedRepository utmGettingStartedRepository;

    public GettingStartedService(UtmGettingStartedRepository utmGettingStartedRepository) {
        this.utmGettingStartedRepository = utmGettingStartedRepository;
    }

    /**
     * initializeSteps If is in SaaS, the omit the APPLICATION_SETTINGS step, this will be managed by us
     *
     * @param isSaas If the instance is running in a SaaS instance or not
     * @throws Exception throws error if there is any issue saving the steps
     */
    public void initializeSteps(boolean isSaas) throws Exception {
        final String ctx = CLASSNAME + ".initializeSteps";
        try {
            utmGettingStartedRepository.deleteAll();
            List<UtmGettingStarted> gettingStartedList = new ArrayList<>();
            for (GettingStartedStepEnum step : GettingStartedStepEnum.values()) {
                // Exclude APPLICATION_SETTINGS when isSaas is true
                //  if (isSaas && step == GettingStartedStepEnum.APPLICATION_SETTINGS) {
                //   continue;
                //  }
                UtmGettingStarted newStep = new UtmGettingStarted();
                newStep.setStepShort(step);
                newStep.setStepOrder(step.ordinal()); // Use ordinal for order
                newStep.setCompleted(false);
                gettingStartedList.add(newStep);

            }
            utmGettingStartedRepository.saveAllAndFlush(gettingStartedList);

        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Set the step status to complete
     *
     * @param stepEnum ShortName enumerator
     * @throws Exception RuntimeException
     */
    public void completeStep(GettingStartedStepEnum stepEnum) throws Exception {
        final String ctx = CLASSNAME + ".completeStep";
        try {
            UtmGettingStarted gettingStarted = utmGettingStartedRepository
                .findByStepShort(stepEnum).orElseThrow(() -> new RuntimeException("Unable to find getting started enum"));
            gettingStarted.setCompleted(true);
            utmGettingStartedRepository.save(gettingStarted);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Get steps
     *
     * @param pageable Pageable
     * @return Page<UtmGettingStarted> Pageable of UtmGettingStarted
     * @throws Exception Exception getting steps
     */
    @Transactional(readOnly = true)
    public Page<UtmGettingStarted> getSteps(Pageable pageable) throws Exception {
        final String ctx = CLASSNAME + ".getSteps";
        try {
            return utmGettingStartedRepository.findAll(pageable);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

}
