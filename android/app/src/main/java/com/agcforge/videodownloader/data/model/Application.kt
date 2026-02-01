package com.agcforge.videodownloader.data.model

import android.os.Parcelable
import com.google.gson.annotations.SerializedName
import kotlinx.parcelize.Parcelize

@Parcelize
data class Application(
    @SerializedName("id") val id: String,
    @SerializedName("name") val name: String,
    @SerializedName("package_name") val packageName: String,
    @SerializedName("version") val version: String? = null,
    @SerializedName("platform") val platform: String = "android",
    @SerializedName("enable_monetization") val enableMonetization: Boolean,
    @SerializedName("enable_admob") val enableAdmob: Boolean,
    @SerializedName("enable_unity_ad") val enableUnityAd: Boolean,
    @SerializedName("enable_start_app") val enableStartApp: Boolean,
    @SerializedName("enable_in_app_purchase") val enableInAppPurchase: Boolean,
    @SerializedName("admob_ad_unit_id") val admobAdUnitId: String? = null,
    @SerializedName("unity_ad_unit_id") val unityAdUnitId: String? = null,
    @SerializedName("start_app_ad_unit_id") val startAppAdUnitId: String? = null,
    @SerializedName("admob_banner_ad_unit_id") val admobBannerAdUnitId: String? = null,
    @SerializedName("admob_interstitial_ad_unit_id") val admobInterstitialAdUnitId: String? = null,
    @SerializedName("admob_native_ad_unit_id") val admobNativeAdUnitId: String? = null,
    @SerializedName("admob_rewarded_ad_unit_id") val admobRewardedAdUnitId: String? = null,
    @SerializedName("unity_banner_ad_unit_id") val unityBannerAdUnitId: String? = null,
    @SerializedName("unity_interstitial_ad_unit_id") val unityInterstitialAdUnitId: String? = null,
    @SerializedName("unity_native_ad_unit_id") val unityNativeAdUnitId: String? = null,
    @SerializedName("unity_rewarded_ad_unit_id") val unityRewardedAdUnitId: String? = null,
    @SerializedName("is_active") val isActive: Boolean,
    @SerializedName("created_at") val createdAt: String,
    @SerializedName("updated_at") val updatedAt: String,
    @SerializedName("in_app_products") val inAppProducts: List<InAppProduct>? = null
) : Parcelable
